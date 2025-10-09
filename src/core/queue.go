package core

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"time"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Queue struct {
	qcount uint

	sbWhiteList *types.SubnetList
	sbBlackList *types.SubnetList
	dmWhiteList *types.DomainList
	dmBlackList *types.DomainList

	filters []filter.Filter
	workers []*Worker
	logger  *logger.Logger
}

func NewQueue(qcount uint, filters []filter.Filter, logger *logger.Logger) *Queue {
	sbWhiteList := types.NewSubnetList()
	sbBlackList := types.NewSubnetList()
	dmWhiteList := types.NewDomainList()
	dmBlackList := types.NewDomainList()

	// prefill subnet whitelist with internal network
	sbWhiteList.Upsert(
		[]netip.Prefix{
			netip.MustParsePrefix("127.0.0.0/8"),
			netip.MustParsePrefix("10.0.0.0/8"),
			netip.MustParsePrefix("192.168.0.0/16"),
			netip.MustParsePrefix("172.16.0.0/12"),
		},
	)

	workers := make([]*Worker, 0, qcount)
	// WARNING: always balancing NFQUEUE from 0
	for qnum := 0; qnum < int(qcount); qnum++ {
		workers = append(workers, NewWorker(uint16(qnum), sbWhiteList, sbBlackList, dmWhiteList, dmBlackList, filters, logger))
	}

	return &Queue{
		qcount:      qcount,
		sbWhiteList: sbWhiteList,
		sbBlackList: sbBlackList,
		dmWhiteList: dmWhiteList,
		dmBlackList: dmBlackList,
		filters:     filters,
		workers:     workers,
		logger:      logger,
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
