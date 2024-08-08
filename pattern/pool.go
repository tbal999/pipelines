package pattern

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
)

type Pool struct {
	wg      *sync.WaitGroup
	options *internalOptions
	success *uint64
	fail    *uint64
	total   *uint64
	ctx     context.Context

	Cache sync.Map
}

// PoolOption functional options within Pool
type PoolOption func(opt *internalOptions) error

type internalOptions struct {
	// the worker that is used in the pool
	worker Worker
	// should we clone the worker object per worker?
	clone bool
	// how many workers do we want in the workerpool?
	workers int
	// the name of the workerpool
	name string
	// the error handler - what we want to do against errors within this worker pool?
	errorHandler func(err error)
	// fail fast? if true - will return on first error
	failFast bool
}

type Worker interface {
	Action(input []byte) ([]byte, error)
	Initialise() error
	Close() error
	Clone() Worker
}

// Name of the Pool
func Name(name string) PoolOption {
	return func(opt *internalOptions) error {
		if strings.TrimSpace(name) == "" {
			return errors.New("must contain a name")
		}

		opt.name = name
		return nil
	}
}

// WorkerCount is how many workers are created
func WorkerCount(workers int) PoolOption {
	return func(opt *internalOptions) error {
		opt.workers = workers
		return nil
	}
}

// WithWorker is how many workers are created
func WithWorker(worker Worker, clone bool) PoolOption {
	return func(opt *internalOptions) error {
		if worker == nil {
			return errors.New("must contain a worker")
		}

		opt.worker = worker

		opt.clone = clone
		return nil
	}
}

// InitFunc is initialize for workerpool
func FailFast() PoolOption {
	return func(opt *internalOptions) error {
		opt.failFast = true
		return nil
	}
}

// WorkerCount is how many workers are created
func ErrorHandler(errorHandler func(err error)) PoolOption {
	return func(opt *internalOptions) error {
		if errorHandler == nil {
			return errors.New("must contain an errorHandler")
		}

		opt.errorHandler = errorHandler
		return nil
	}
}

func NewPool(ctx context.Context, opts ...PoolOption) (Pool, error) {
	p := Pool{}

	var success, fail, total uint64

	p.success = &success
	p.fail = &fail
	p.total = &total

	p.options = &internalOptions{}

	for _, opt := range opts {
		err := opt(p.options)
		if err != nil {
			return p, err
		}
	}

	p.wg = &sync.WaitGroup{}

	p.Cache = sync.Map{}

	p.ctx = ctx

	return p, nil
}

func (p *Pool) TotalSuccess() uint64 {
	count := atomic.LoadUint64(p.success)
	return count
}

func (p *Pool) TotalFail() uint64 {
	count := atomic.LoadUint64(p.fail)
	return count
}

func (p *Pool) Total() uint64 {
	count := atomic.LoadUint64(p.total)
	return count
}

func (p *Pool) Name() string {
	return p.options.name
}

func (p *Pool) Start(inputChan chan []byte, outputChans []chan []byte) {
	workers := []Worker{}

	defer func() {
		for index := range outputChans {
			close(outputChans[index])
		}
	}()

	for i := 0; i < p.options.workers; i++ {
		if p.options.clone {
			clonedWorker := p.options.worker.Clone()

			workers = append(workers, clonedWorker)

			p.wg.Add(1)

			go func() {
				defer p.wg.Done()
				for x := range inputChan {
					result, err := clonedWorker.Action(x)
					if err != nil {
						p.options.errorHandler(fmt.Errorf("workerpool [%s]: %w", p.options.name, err))
						atomic.AddUint64(p.fail, 1)
						if p.options.failFast {
							return
						}
					} else {
						for index := range outputChans {
							outputChans[index] <- result
						}

						atomic.AddUint64(p.success, 1)
					}

					atomic.AddUint64(p.total, 1)
				}
			}()
		} else {
			workers = append(workers, p.options.worker)

			p.wg.Add(1)

			go func() {
				defer p.wg.Done()
				for x := range inputChan {
					result, err := p.options.worker.Action(x)
					if err != nil {
						p.options.errorHandler(fmt.Errorf("workerpool [%s]: %w", p.options.name, err))
						atomic.AddUint64(p.fail, 1)
						if p.options.failFast {
							return
						}
					} else {
						for index := range outputChans {
							outputChans[index] <- result
						}

						atomic.AddUint64(p.success, 1)
					}

					atomic.AddUint64(p.total, 1)
				}
			}()
		}
	}

	p.wg.Wait()

	for index := range workers {
		_ = workers[index].Close()
	}

	log.Println("done ", p.options.name)
}
