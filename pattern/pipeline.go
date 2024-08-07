package pattern

type PoolMap struct {
	Instance *Pool
	Children []PoolMap
	Channels []chan []byte
}

func (p *PoolMap) GetChannels() []chan []byte {
	return p.Channels
}

func Initialise(workerPools *PoolMap) {
	numChannels := len(workerPools.Children)

	if numChannels == 0 {
		numChannels = 1
	}

	workerPools.Channels = make([]chan []byte, numChannels)

	for i := range workerPools.Channels {
		workerPools.Channels[i] = make(chan []byte)
	}

	for index := range workerPools.Children {
		Initialise(&workerPools.Children[index])
	}
}

func Start(workerPools *PoolMap, inputChan chan []byte) {
	if workerPools.Instance != nil {
		go workerPools.Instance.Start(inputChan, workerPools.Channels)
	}

	for index := range workerPools.Children {
		go Start(&workerPools.Children[index], workerPools.Channels[index])
	}
}
