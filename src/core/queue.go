package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/appleboy/graceful"
)

type Queue struct {
	qcount  uint
	workers []*Worker
	logger  graceful.Logger
}

func NewQueue(qcount uint, filters []Filter, logger graceful.Logger) *Queue {
	workers := make([]*Worker, 0, qcount)
	for qnum := 0; qnum < int(qcount); qnum++ {
		workers = append(workers, NewWorker(uint16(qnum), filters, logger))
	}

	return &Queue{
		qcount:  qcount,
		workers: workers,
		logger:  logger,
	}
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
