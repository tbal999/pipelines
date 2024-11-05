package pkg

import (
	"sync"
)

// PoolTree structure is a tree of workerpools
// that feed downwards to form a pipeline
type PoolTree struct {
	WorkerPool *Pool
	PublishTo  []PoolTree
}

func (workerPools *PoolTree) initialise() {
	if workerPools.WorkerPool == nil {
		return
	}

	workerPools.WorkerPool.Lock()
	defer workerPools.WorkerPool.Unlock()

	numChannels := len(workerPools.PublishTo)

	workerPools.WorkerPool.Channels = make([]chan []byte, numChannels)

	for i := range workerPools.WorkerPool.Channels {
		if workerPools.WorkerPool.BufferSize() == 0 {
			workerPools.WorkerPool.Channels[i] = make(chan []byte)
		} else {
			workerPools.WorkerPool.Channels[i] = make(chan []byte, workerPools.WorkerPool.BufferSize())
		}
	}

	for index := range workerPools.PublishTo {
		workerPools.PublishTo[index].initialise()
	}
}

func (workerPools *PoolTree) Start(inputChan <-chan []byte) {
	workerPools.initialise()

	wg := &sync.WaitGroup{}

	if workerPools.WorkerPool != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()
			workerPools.WorkerPool.Start(inputChan)
		}()

		for index := range workerPools.PublishTo {
			wg.Add(1)

			go func(i int) {
				defer wg.Done()
				workerPools.PublishTo[i].Start(workerPools.WorkerPool.Channels[i])
			}(index)
		}

		wg.Wait()
	}
}
