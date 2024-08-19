package workers

import (
	"log"

	pipeline "github.com/tbal999/pipelines/pkg"
)

type Mapper struct {
	internalCounter int
	Log             bool
}

func (w *Mapper) Clone() pipeline.Worker {
	clone := *w

	return &clone
}

func (w *Mapper) Initialise(configBytes []byte) error {
	log.Println("mapper started")

	return nil
}

func (w *Mapper) Close() error {
	log.Println("mapper stopped")
	return nil
}

func (w *Mapper) Action(input []byte) ([]byte, error) {
	// to demonstrate concurrency safe even without mutex
	w.internalCounter += 1

	if w.Log {
		log.Println(string(input))
	}

	return input, nil
}
