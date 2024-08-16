package pattern

//go:generate mockgen --build_flags=--mod=mod -destination=./mocks/worker.go -package=mocks github.com/tbal999/pipelines/pattern Worker

type Worker interface {
	Action(input []byte) ([]byte, error)
	Initialise(configBytes []byte) error
	Close() error
	Clone() Worker
}
