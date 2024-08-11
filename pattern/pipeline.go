package pattern

import "sync"

// PoolTree structure is a tree of workerpools
// that feed downwards to form a pipeline
type PoolTree struct {
	WorkerPool *Pool
	Children   []PoolTree
}

func (workerPools *PoolTree) initialise() {
	workerPools.WorkerPool.Lock()
	defer workerPools.WorkerPool.Unlock()

	numChannels := len(workerPools.Children)

	if numChannels == 0 {
		// there will always be at least one output channel
		// to enable bespoke handling at egress (if needed)
		numChannels = 1
	}

	workerPools.WorkerPool.Channels = make([]chan []byte, numChannels)

	for i := range workerPools.WorkerPool.Channels {
		if workerPools.WorkerPool.BufferSize() == 0 {
			workerPools.WorkerPool.Channels[i] = make(chan []byte)
		} else {
			workerPools.WorkerPool.Channels[i] = make(chan []byte, workerPools.WorkerPool.BufferSize())
		}
	}

	for index := range workerPools.Children {
		workerPools.Children[index].initialise()
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

		for index := range workerPools.Children {
			wg.Add(1)

			go func() {
				defer wg.Done()
				workerPools.Children[index].Start(workerPools.WorkerPool.Channels[index])
			}()
		}
	}

	wg.Wait()
}
