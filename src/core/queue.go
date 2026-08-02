package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
	"github.com/cnaize/meds/src/core/metrics"
)

const MedsChainName = "MEDS"

const (
	ConnMarkBlockList uint32 = 0x100000 << iota
	ConnMarkTrustList
)

type Queue struct {
	cfg *config.Config

	logger  *logger.Logger
	filters []filter.Filter

	readers []*Reader
	workers []*Worker
}

func NewQueue(cfg *config.Config, filters []filter.Filter, logger *logger.Logger) *Queue {
	readers := make([]*Reader, 0, cfg.ReadersCount)
	workers := make([]*Worker, 0, cfg.ReadersCount*cfg.WorkersCount)
	// WARNING: always balancing NFQUEUE from 0
	for qnum := 0; qnum < int(cfg.ReadersCount); qnum++ {
		reader := NewReader(uint16(qnum), uint32(cfg.ReaderQLen), cfg.AcceptOnFail, logger)
		readers = append(readers, reader)

		// workers per reader
		for range cfg.WorkersCount {
			workers = append(workers, NewWorker(filters, logger))
		}
	}

	return &Queue{
		cfg:     cfg,
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
		for j := i * int(q.cfg.WorkersCount); j < i*int(q.cfg.WorkersCount)+int(q.cfg.WorkersCount); j++ {
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
	if err := q.iptablesUp(); err != nil {
		return fmt.Errorf("iptables up: %w", err)
	}

	// wait till the end
	<-ctx.Done()
	return nil
}

func (q *Queue) Update(ctx context.Context, timeout, interval time.Duration) {
	ticker := time.NewTicker(interval)
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

		// wait
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
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
	if err := q.iptablesDown(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("iptables down: %w", err))
	}

	return errs
}

func (q *Queue) iptablesUp() error {
	blockListMark := "0x" + strconv.FormatUint(uint64(ConnMarkBlockList), 16)
	trustListMark := "0x" + strconv.FormatUint(uint64(ConnMarkTrustList), 16)

	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("iptables new: %w", err)
	}

	if err := manageMarkRules(ipt.AppendUnique); err != nil {
		return fmt.Errorf("manage mark rules: %w", err)
	}

	if ok, err := ipt.ChainExists("filter", MedsChainName); err != nil {
		return fmt.Errorf("chain exists: %w", err)
	} else if !ok {
		if err := ipt.NewChain("filter", MedsChainName); err != nil {
			return fmt.Errorf("new chain: %w", err)
		}
	}

	if err := ipt.AppendUnique(
		"filter",
		MedsChainName,
		"-m",
		"mark",
		"--mark",
		blockListMark+"/"+blockListMark,
		"-j",
		"DROP",
	); err != nil {
		return err
	}

	medsArgs := []string{
		"-m",
		"mark",
		"!",
		"--mark",
		trustListMark,
		"-m",
		"connbytes",
		"--connbytes-mode",
		"packets",
		"--connbytes",
		"0:10",
		"--connbytes-dir",
		"original",
		"-j",
		"NFQUEUE",
		"--queue-bypass",
	}
	if q.cfg.ReadersCount > 1 {
		medsArgs = append(medsArgs, "--queue-balance", fmt.Sprintf("0:%d", q.cfg.ReadersCount-1))
	}
	if err := ipt.AppendUnique("filter", MedsChainName, medsArgs...); err != nil {
		return err
	}

	return ipt.AppendUnique("filter", "INPUT", "-j", MedsChainName)
}

func (q *Queue) iptablesDown() error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("iptables new: %w", err)
	}

	if err := manageMarkRules(ipt.DeleteIfExists); err != nil {
		return fmt.Errorf("manage mark rules: %w", err)
	}

	if err := ipt.DeleteIfExists("filter", "INPUT", "-j", MedsChainName); err != nil {
		return err
	}

	return ipt.ClearAndDeleteChain("filter", MedsChainName)
}

func manageMarkRules(action func(table, chain string, rulespec ...string) error) error {
	if err := action(
		"mangle",
		"PREROUTING",
		"-m",
		"comment",
		"--comment",
		MedsChainName,
		"-j",
		"CONNMARK",
		"--restore-mark",
		"--mask",
		"0xFFFFFFFF",
	); err != nil {
		return err
	}

	return action(
		"mangle",
		"POSTROUTING",
		"-m",
		"comment",
		"--comment",
		MedsChainName,
		"-j",
		"CONNMARK",
		"--save-mark",
		"--mask",
		"0xFFFFFFFF",
	)
}
