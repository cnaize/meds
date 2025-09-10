package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/appleboy/graceful"
	"golang.org/x/sync/errgroup"

	"github.com/cnaize/meds/src/core/filter"
)

type Queue struct {
	qcount  uint
	filters []filter.Filter
	workers []*Worker
	logger  graceful.Logger
}

func NewQueue(qcount uint, filters []filter.Filter, logger graceful.Logger) *Queue {
	workers := make([]*Worker, 0, qcount)
	// WARNING: always balancing from 0 to qcount
	for qnum := 0; qnum < int(qcount); qnum++ {
		workers = append(workers, NewWorker(uint16(qnum), filters, logger))
	}

	return &Queue{
		qcount:  qcount,
		filters: filters,
		workers: workers,
		logger:  logger,
	}
}

func (q *Queue) Load(ctx context.Context) error {
	q.logger.Infof("Loading queue...")

	group, ctx := errgroup.WithContext(ctx)
	for i, filter := range q.filters {
		group.Go(func() error {
			if err := filter.Load(ctx); err != nil {
				return fmt.Errorf("%d: load filter: %w", i, err)
			}

			return nil
		})

	}
	return group.Wait()
}

func (q *Queue) Run(ctx context.Context) error {
	q.logger.Infof("Running queue...")

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

func (q *Queue) Update(ctx context.Context, interval time.Duration) {
	for {
		q.logger.Infof("Updating queue...")

		// update filters
		var wg sync.WaitGroup
		for i, filter := range q.filters {
			wg.Go(func() {
				if err := filter.Update(ctx); err != nil {
					q.logger.Errorf("%d: failed to update filter: %s", i, err.Error())
				}
			})
		}
		wg.Wait()

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
	out, err := exec.Command(
		"iptables",
		"-I",
		"INPUT",
		"-j",
		"NFQUEUE",
		"--queue-balance",
		fmt.Sprintf("%d:%d", 0, q.qcount), // WARNING: always balancing from 0 to qcount
		"--queue-bypass").
		CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec: %s", out)
	}

	return nil
}

func (q *Queue) ipTablesDown() error {
	out, err := exec.Command(
		"iptables",
		"-D",
		"INPUT",
		"-j",
		"NFQUEUE",
		"--queue-balance",
		fmt.Sprintf("%d:%d", 0, q.qcount), // WARNING: always balancing from 0 to qcount
		"--queue-bypass").
		CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec: %s", out)
	}

	return nil
}
