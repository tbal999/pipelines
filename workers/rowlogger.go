package workers

import (
	"log"

	"time"

	pipeline "github.com/tbal999/pipelines/pattern"
)

type RowLogger struct {
}

func (w *RowLogger) Clone() pipeline.Worker {
	clone := *w

	return &clone
}

func (w *RowLogger) Initialise(configBytes []byte) error {
	return nil
}

func (w *RowLogger) Close() error {
	return nil
}

func (w *RowLogger) Action(input []byte) ([]byte, error) {
	time.Sleep(1 * time.Second)
	log.Println(string(input))
	return input, nil
	/*
		// to test thread safety without mutexes
		n.internalValue = n.internalValue + 1

		type Input struct {
			Number int
		}

		var item Input

		_ = json.Unmarshal(input, &item)

		item.Number = item.Number * n.Multiplier

		return json.Marshal(item)*/
}
