package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Queue struct {
	qcount uint
	logger *logger.Logger

	filters []filter.Filter
	workers []*Worker
}

func NewQueue(
	qcount uint,
	wqlen uint,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
	filters []filter.Filter,
	logger *logger.Logger,
) *Queue {
	workers := make([]*Worker, 0, qcount)
	// WARNING: always balancing NFQUEUE from 0
	for qnum := 0; qnum < int(qcount); qnum++ {
		workers = append(workers,
			NewWorker(uint16(qnum), uint32(wqlen), subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList, filters, logger),
		)
	}

	return &Queue{
		qcount:  qcount,
		logger:  logger,
		filters: filters,
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

	// create workers
	for _, worker := range q.workers {
		if err := worker.Run(ctx); err != nil {
			return fmt.Errorf("%d: worker run: %w", worker.qnum, err)
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
					q.logger.Raw().
						Error().
						Err(err).
						Str("name", filter.Name()).
						Str("type", string(filter.Type())).
						Msg("filter update failed")
				}
			}()
		}

		// sleep
		time.Sleep(interval)
	}
}

func (q *Queue) Close() error {
	var errs error
	// down iptables
	if err := q.ipTablesDown(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("iptables down: %w", err))
	}

	// close workers
	for _, worker := range q.workers {
		if err := worker.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
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
