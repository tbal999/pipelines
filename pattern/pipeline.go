package pattern

type PoolMap struct {
	channels []chan []byte

	Instance *Pool
	Children []PoolMap
}

func (p *PoolMap) GetChannels() []chan []byte {
	return p.channels
}

func Initialise(workerPools *PoolMap) {
	numChannels := len(workerPools.Children)

	if numChannels == 0 {
		numChannels = 1
	}

	workerPools.channels = make([]chan []byte, numChannels)

	for i := range workerPools.channels {
		workerPools.channels[i] = make(chan []byte)
	}

	for index := range workerPools.Children {
		Initialise(&workerPools.Children[index])
	}
}

func Start(workerPools *PoolMap, inputChan chan []byte) {
	if workerPools.Instance != nil {
		go workerPools.Instance.Start(inputChan, workerPools.channels)
	}

	for index := range workerPools.Children {
		go Start(&workerPools.Children[index], workerPools.channels[index])
	}
}
