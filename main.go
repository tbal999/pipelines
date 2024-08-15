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
	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	mappingWorkerPool, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 1"),
		pattern.WorkerCount(2),
		pattern.BufferSize(4),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte("$$")),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	mappingWorkerPool2, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 2"),
		pattern.WorkerCount(2),
		pattern.BufferSize(4),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte("[$$]")),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	mappingWorkerPool3, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 3"),
		pattern.WorkerCount(2),
		pattern.BufferSize(4),
		pattern.Final(),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte(`$$[0].event`)),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	mappingWorkerPool4, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool 4"),
		pattern.WorkerCount(2),
		pattern.BufferSize(4),
		pattern.Final(),
		pattern.WithWorker(&workers.Mapper{Log: true}),
		pattern.WorkerConfigBytes([]byte(`$$[0].event`)),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	inputBytes := readers.SendEvents(ctx, 500000)

	// a tree of worker pools that form a pipeline
	// 1 sends to 2 
	// 2 sends to 3 and 4
	workerPipeline := pattern.PoolTree{
		WorkerPool: mappingWorkerPool,
		PublishTo: []pattern.PoolTree{
			{
				WorkerPool: mappingWorkerPool2,
				PublishTo: []pattern.PoolTree{
					{
						WorkerPool: mappingWorkerPool3,
					},
					{
						WorkerPool: mappingWorkerPool4,
					},
				},
			},
		},
	}

	workerPipeline.Init()

	workerPipeline.Start(inputBytes)
}
