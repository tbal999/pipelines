package pkg

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
	// how many workers do we want in the workerpool?
	workers int
	// the name of the workerpool
	name string
	// the error handler - what we want to do against errors within this worker pool?
	errorHandler func(err error)

	end bool

	buffer int
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

// BufferSize is the size of the egress channels buffer
func BufferSize(buffer int) PoolOption {
	return func(opt *internalOptions) error {
		opt.buffer = buffer
		return nil
	}
}

// NoOutput forces the workerpool to drain it's output channels automatically
// i.e it doesn't have an egress
func Final() PoolOption {
	return func(opt *internalOptions) error {
		opt.end = true
		return nil
	}
}

// WorkerConfigBytes is whatever the config is needed for the worker to work
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

func (p *Pool) BufferSize() int {
	return p.options.buffer
}

func (p *Pool) Lock() {
	p.mu.Lock()
}

func (p *Pool) Unlock() {
	p.mu.Unlock()
}

func (p *Pool) Start(inputChan <-chan []byte) {
	p.Lock()
	defer p.Unlock()

	wg := &sync.WaitGroup{}

	for i := 0; i < p.options.workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			p.startWorker(inputChan)
		}()
	}

	wg.Wait()

	for index := range p.Channels {
		close(p.Channels[index])
	}
}

func (p *Pool) startWorker(inputChan <-chan []byte) {
	clonedWorker := p.options.worker.Clone()

	err := clonedWorker.Initialise(p.options.configBytes)
	if err != nil {
		p.options.errorHandler(fmt.Errorf("workerpool init error [%s]: %w", p.options.name, err))
		return
	}

	defer func() {
		err := clonedWorker.Close()
		if err != nil {
			p.options.errorHandler(fmt.Errorf("workerpool close error [%s]: %w", p.options.name, err))
		}
	}()

	for x := range inputChan {
		select {
		case <-p.ctx.Done():
			for range inputChan {
				// drain and finish
			}

			return
		default:
			// continue processing
		}

		result, send, err := clonedWorker.Action(x)
		if err != nil {
			p.options.errorHandler(fmt.Errorf("workerpool action error [%s]: %w", p.options.name, err))

			atomic.AddUint64(p.fail, 1)
		} else {
			if !p.options.end && send {
				for index := range p.Channels {
					p.Channels[index] <- result
				}
			}

			atomic.AddUint64(p.success, 1)
		}

		atomic.AddUint64(p.total, 1)
	}
}
