package main

import (
	"context"
	"log"
	"time"

	"github.com/tbal999/pipelines/pattern"
	"github.com/tbal999/pipelines/readers"
	"github.com/tbal999/pipelines/workers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		time.Sleep(300 * time.Second)
		cancel()
	}()

	mappingWorkerPool, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 1"),
		pattern.WorkerCount(1),
		pattern.BufferSize(1),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte("$$")),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
			//cancel()
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	mappingWorkerPool2, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 2"),
		pattern.WorkerCount(1),
		pattern.BufferSize(1),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte("[$$]")),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
			//cancel()
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	mappingWorkerPool3, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 3"),
		pattern.WorkerCount(6),
		pattern.BufferSize(1),
		pattern.Final(),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte(`$$[0].Name`)),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
			//cancel()
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	inputBytes := readers.HTTPEvents()

	// a tree of worker pools that form a pipeline
	workerPipeline := pattern.PoolTree{
		WorkerPool: mappingWorkerPool,
		Children: []pattern.PoolTree{
			{
				WorkerPool: mappingWorkerPool2,
				Children: []pattern.PoolTree{
					{
						WorkerPool: mappingWorkerPool3,
					},
				},
			},
		},
	}

	workerPipeline.Start(inputBytes)
}
