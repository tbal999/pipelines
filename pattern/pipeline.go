package pattern

import "sync"

type PoolTree struct {
	Instance *Pool
	Children []PoolTree
}

func (workerPools *PoolTree) initialise() {
	workerPools.Instance.Lock()
	defer workerPools.Instance.Unlock()

	numChannels := len(workerPools.Children)

	if numChannels == 0 {
		// there will always be at least one output channel
		// to enable bespoke handling at egress (if needed)
		numChannels = 1
	}

	workerPools.Instance.Channels = make([]chan []byte, numChannels)

	for i := range workerPools.Instance.Channels {
		if workerPools.Instance.BufferSize() == 0 {
			workerPools.Instance.Channels[i] = make(chan []byte)
		} else {
			workerPools.Instance.Channels[i] = make(chan []byte, workerPools.Instance.BufferSize())
		}
	}

	for index := range workerPools.Children {
		workerPools.Children[index].initialise()
	}
}

func (workerPools *PoolTree) Start(inputChan <-chan []byte) {
	workerPools.initialise()

	wg := &sync.WaitGroup{}

	if workerPools.Instance != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerPools.Instance.Start(inputChan)
		}()
	}

	for index := range workerPools.Children {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerPools.Children[index].Start(workerPools.Instance.Channels[index])
		}()
	}

	wg.Wait()
}
