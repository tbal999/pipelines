package pattern

func Start(workerPools []Pool, inputChan chan []byte) <-chan []byte {
	numChannels := len(workerPools) + 1

	channels := make([]chan []byte, numChannels)

	for i := range channels {
		channels[i] = make(chan []byte)
	}

	channels[0] = inputChan

	for index := range workerPools {
		go workerPools[index].Start(channels[index], channels[index+1])
	}

	// this channel will be gracefully closed by the last item in the pipeline
	return channels[len(channels)-1]
}
