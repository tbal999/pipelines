package workers

import (
	"log"
    "gopkg.in/yaml.v2"
	pipeline "github.com/tbal999/pipelines/pkg"
)

type Logger struct {
	internalCounter int
	Log             bool
}

func (w *Logger) Clone() pipeline.Worker {
	clone := *w

	return &clone
}

func (w *Logger) Initialise(configBytes []byte) error {
	return yaml.Unmarshal(configBytes, &w)
}

func (w *Logger) Close() error {
	log.Println("logger stopped")
	return nil
}

func (w *Logger) Action(input []byte) ([]byte, bool, error) {
	// to demonstrate concurrency safe even without mutex
	w.internalCounter += 1

	if w.Log {
		log.Println(string(input), w.internalCounter)
	}

	return input, true, nil
}
