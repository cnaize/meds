package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/types"
)

type Queue struct {
	qcount uint
	wcount uint

	logger  *logger.Logger
	filters []filter.Filter

	readers []*Reader
	workers []*Worker
}

func NewQueue(
	qcount uint,
	wcount uint,
	qlen uint,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
	filters []filter.Filter,
	logger *logger.Logger,
) *Queue {
	readers := make([]*Reader, 0, qcount)
	workers := make([]*Worker, 0, qcount*wcount)
	// WARNING: always balancing NFQUEUE from 0
	for qnum := 0; qnum < int(qcount); qnum++ {
		reader := NewReader(uint16(qnum), uint32(qlen), logger)

		// workers per reader
		for range wcount {
			workers = append(workers,
				NewWorker(
					subnetWhiteList,
					subnetBlackList,
					domainWhiteList,
					domainBlackList,
					filters,
					logger,
				),
			)
		}

		readers = append(readers, reader)
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
	cmd := exec.Command("iptables", "-I", "INPUT", "-j", "NFQUEUE", "--queue-bypass")
	if q.qcount > 1 {
		// WARNING: always balancing NFQUEUE from 0
		cmd.Args = append(cmd.Args, "--queue-balance", fmt.Sprintf("%d:%d", 0, q.qcount-1))
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec: %s", out)
	}

	return nil
}

func (q *Queue) ipTablesDown() error {
	cmd := exec.Command("iptables", "-D", "INPUT", "-j", "NFQUEUE", "--queue-bypass")
	if q.qcount > 1 {
		// WARNING: always balancing NFQUEUE from 0
		cmd.Args = append(cmd.Args, "--queue-balance", fmt.Sprintf("%d:%d", 0, q.qcount-1))
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec: %s", out)
	}

	return nil
}
