package pkg

//go:generate mockgen --build_flags=--mod=mod -destination=./mocks/worker.go -package=mocks github.com/tbal999/pipelines/pkg Worker

type Worker interface {
	Action(input []byte) ([]byte, error)
	Initialise(configBytes []byte) error
	Close() error
	Clone() Worker
}
