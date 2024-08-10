package pattern

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type Pool struct {
	mu      *sync.Mutex
	options *internalOptions
	success *uint64
	fail    *uint64
	total   *uint64
	ctx     context.Context

	Channels []chan []byte
}

func (p *Pool) GetOutputChannels() []chan []byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Channels
}

// PoolOption functional options within Pool
type PoolOption func(opt *internalOptions) error

type internalOptions struct {
	// the worker that is used in the pool
	worker Worker
	// the config bytes used by the worker
	configBytes []byte
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
	Initialise(configBytes []byte) error
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

// WorkerCount is how many workers are created
func WorkerConfigBytes(workerConfig []byte) PoolOption {
	return func(opt *internalOptions) error {
		opt.configBytes = workerConfig
		return nil
	}
}

// WithWorker is how many workers are created
func WithWorker(worker Worker) PoolOption {
	return func(opt *internalOptions) error {
		if worker == nil {
			return errors.New("must contain a worker")
		}

		opt.worker = worker

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

func NewPool(ctx context.Context, opts ...PoolOption) (*Pool, error) {
	p := Pool{}

	var success, fail, total uint64

	p.success = &success
	p.fail = &fail
	p.total = &total
	p.ctx = ctx
	p.mu = &sync.Mutex{}

	p.options = &internalOptions{}

	for _, opt := range opts {
		err := opt(p.options)
		if err != nil {
			return &p, err
		}
	}

	return &p, nil
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

func (p *Pool) Start(inputChan <-chan []byte) {
	wg := &sync.WaitGroup{}

	workers := []Worker{}

	for i := 0; i < p.options.workers; i++ {
		clonedWorker := p.options.worker.Clone()

		err := clonedWorker.Initialise(p.options.configBytes)
		if err != nil {
			p.options.errorHandler(fmt.Errorf("workerpool init error [%s]: %w", p.options.name, err))
			continue
		}

		workers = append(workers, clonedWorker)

		wg.Add(1)

		go func() {
			defer wg.Done()
			for x := range inputChan {
				result, err := clonedWorker.Action(x)
				if err != nil {
					p.options.errorHandler(fmt.Errorf("workerpool action error [%s]: %w", p.options.name, err))

					atomic.AddUint64(p.fail, 1)

					if p.options.failFast {
						return
					}
				} else {
					for index := range p.Channels {
						p.Channels[index] <- result
					}

					atomic.AddUint64(p.success, 1)
				}

				atomic.AddUint64(p.total, 1)
			}
		}()
	}

	wg.Wait()

	for index := range workers {
		err := workers[index].Close()
		if err != nil {
			p.options.errorHandler(fmt.Errorf("workerpool close error [%s]: %w", p.options.name, err))
		}
	}

	for index := range p.Channels {
		close(p.Channels[index])
	}
}
