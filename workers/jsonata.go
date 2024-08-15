package workers

import (
	"fmt"
	"log"
	"sync"

	"github.com/xiatechs/jsonata-go"

	pipeline "github.com/tbal999/pipelines/pattern"
)

// cache is used to store the JSONata mapper file on first read and then used by all future
// JSONata mapper initialisations.
var cache sync.Map

type Mapper struct {
	parser *jsonata.Expr

	Log bool
}

func (w *Mapper) Clone() pipeline.Worker {
	clone := *w

	return &clone
}

func (w *Mapper) Initialise(configBytes []byte) error {
	if cachedParser, ok := cache.Load(string(configBytes)); ok {
		w.parser, ok = cachedParser.(*jsonata.Expr)
		if !ok {
			return fmt.Errorf("%w : %s is not a valid jsonata expression", "not valid", string(configBytes))
		}

		return nil
	}

	parser, err := jsonata.Compile(string(configBytes))
	if err != nil {
		return err
	}

	cache.Store(string(configBytes), parser)

	w.parser = parser

	log.Println("mapper started")

	return nil
}

func (w *Mapper) Close() error {
	log.Println("mapper stopped")
	return nil
}

func (w *Mapper) Action(input []byte) ([]byte, error) {
	result, err := w.parser.EvalBytes(input)
	if err != nil {
		return nil, err
	}

	if w.Log {
		log.Println(string(result))
	}

	return result, nil
}
