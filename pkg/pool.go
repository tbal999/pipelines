package pattern

import (
	"errors"
	"fmt"
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
}

// PoolOption functional options within Pool
type PoolOption func(opt *internalOptions) error

type internalOptions struct {
	// how many workers do we want in the workerpool?
	workers int
	// the name of the workerpool
	name string
	// the function used against the []byte data - this is done at a worker level
	workerFunc func(input []byte) ([]byte, error)
	// the error handler - what we want to do against errors within this worker pool?
	errorHandler func(err error)
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
func Function(workerFunc func(input []byte) ([]byte, error)) PoolOption {
	return func(opt *internalOptions) error {
		if workerFunc == nil {
			return errors.New("must contain a workerFunc")
		}

		opt.workerFunc = workerFunc
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

func NewPool(opts ...PoolOption) (Pool, error) {
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

func (p *Pool) Start(inputChan chan []byte, outputChan chan<- []byte) {
	for i := 0; i < p.options.workers; i++ {
		p.wg.Add(1)

		go func() {
			defer p.wg.Done()
			for x := range inputChan {
				result, err := p.options.workerFunc(x)
				if err != nil {
					p.options.errorHandler(fmt.Errorf("workerpool [%s]: %w", p.options.name, err))
					atomic.AddUint64(p.fail, 1)
				} else {
					outputChan <- result
					atomic.AddUint64(p.success, 1)
				}

				atomic.AddUint64(p.total, 1)
			}
		}()
	}

	p.wg.Wait()

	close(outputChan)
}
