package pattern

import (
	"log"
)

type PoolTree struct {
	Instance *Pool
	Children []PoolTree
}

func (workerPools *PoolTree) Initialise() {
	workerPools.Instance.Lock()
	defer workerPools.Instance.Unlock()
	numChannels := len(workerPools.Children)

	if numChannels == 0 {
		// there should always be at least one output channel
		numChannels = 1
	}

	workerPools.Instance.Channels = make([]chan []byte, numChannels)

	for i := range workerPools.Instance.Channels {
		workerPools.Instance.Channels[i] = make(chan []byte)
	}

	for index := range workerPools.Children {
		workerPools.Children[index].Initialise()
	}
}

func (workerPools *PoolTree) Start(inputChan <-chan []byte) {

	if workerPools.Instance != nil {
		go workerPools.Instance.Start(inputChan)
	}

	log.Println(workerPools.Instance.Name(), len(workerPools.Instance.Channels))

	for index := range workerPools.Children {
		go workerPools.Children[index].Start(workerPools.Instance.Channels[index])
	}
}
