package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
	"github.com/cnaize/meds/src/core/metrics"
)

const ConnMark uint32 = 0x100000

type Queue struct {
	qcount uint
	wcount uint

	logger  *logger.Logger
	filters []filter.Filter

	readers []*Reader
	workers []*Worker
}

func NewQueue(qcount uint, wcount uint, qlen uint, filters []filter.Filter, logger *logger.Logger) *Queue {
	readers := make([]*Reader, 0, qcount)
	workers := make([]*Worker, 0, qcount*wcount)
	// WARNING: always balancing NFQUEUE from 0
	for qnum := 0; qnum < int(qcount); qnum++ {
		reader := NewReader(uint16(qnum), uint32(qlen), logger)
		readers = append(readers, reader)

		// workers per reader
		for range wcount {
			workers = append(workers, NewWorker(filters, logger))
		}
	}

	return &Queue{
		qcount:  qcount,
		wcount:  wcount,
		logger:  logger,
		filters: filters,
		readers: readers,
		workers: workers,
	}
}

func (q *Queue) Load(ctx context.Context) error {
	q.logger.Raw().Info().Msg("Loading queue...")

	for _, filter := range q.filters {
		if err := filter.Load(ctx); err != nil {
			return fmt.Errorf("%s (%s): filter load: %w", filter.Name(), filter.Type(), err)
		}
	}

	return nil
}

func (q *Queue) Run(ctx context.Context) error {
	q.logger.Raw().Info().Msg("Running queue...")

	// run readers
	for i, reader := range q.readers {
		if err := reader.Run(ctx); err != nil {
			return fmt.Errorf("%d: reader run: %w", reader.qnum, err)
		}

		// run workers
		for j := i * int(q.wcount); j < i*int(q.wcount)+int(q.wcount); j++ {
			go func() {
				if err := q.workers[j].Run(ctx, reader.nfq, reader.wch); err != nil {
					msg := "worker run"

					metrics.Get().ErrorsTotal.WithLabelValues(msg).Inc()
					q.logger.Log(event.NewError(zerolog.ErrorLevel, msg, err))
				}
			}()
		}
	}

	// up iptables
	if err := q.ipTablesUp(); err != nil {
		return fmt.Errorf("iptables up: %w", err)
	}

	// wait till the end
	<-ctx.Done()
	return nil
}

func (q *Queue) Update(ctx context.Context, timeout, interval time.Duration) {
	for {
		q.logger.Raw().Info().Msg("Updating queue...")

		// update filters
		for _, filter := range q.filters {
			func() {
				// timeout is per filter
				ctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				if err := filter.Update(ctx); err != nil {
					msg := "filter update failed"

					metrics.Get().ErrorsTotal.WithLabelValues(msg).Inc()
					q.logger.Raw().
						Error().
						Err(err).
						Str("name", filter.Name()).
						Str("type", string(filter.Type())).
						Msg(msg)
				}
			}()
		}

		// sleep
		time.Sleep(interval)
	}
}

func (q *Queue) Close() error {
	var errs error
	// close readers
	for _, reader := range q.readers {
		if err := reader.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("reader close: %w", err))
		}
	}

	// down iptables
	if err := q.ipTablesDown(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("iptables down: %w", err))
	}

	return errs
}

func (q *Queue) ipTablesUp() error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("iptables new: %w", err)
	}

	return q.manageIptables(ipt.AppendUnique)
}

func (q *Queue) ipTablesDown() error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("iptables new: %w", err)
	}

	return q.manageIptables(ipt.DeleteIfExists)
}

func (q *Queue) manageIptables(action func(table, chain string, rulespec ...string) error) error {
	mark := "0x" + strconv.FormatUint(uint64(ConnMark), 16)
	comment := "MEDS_NET_HEALING"

	if err := action("mangle", "PREROUTING", "-j", "CONNMARK", "--restore-mark", "--mask", mark, "-m", "comment", "--comment", comment); err != nil {
		return err
	}

	if err := action("filter", "INPUT", "-m", "connmark", "--mark", mark+"/"+mark, "-m", "comment", "--comment", comment, "-j", "ACCEPT"); err != nil {
		return err
	}

	args := []string{"-m", "connmark", "--mark", "0x0/" + mark, "-m", "comment", "--comment", comment, "-j", "NFQUEUE", "--queue-bypass"}
	if q.qcount > 1 {
		args = append(args, "--queue-balance", fmt.Sprintf("0:%d", q.qcount-1))
	}

	return action("filter", "INPUT", args...)
}
