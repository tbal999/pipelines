package main

import (
	"context"
	"log"
	"time"

	"github.com/tbal999/pipelines/pattern"
	"github.com/tbal999/pipelines/workers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		time.Sleep(60 * time.Second)
	}()

	mappingWorkerPool, err := pattern.NewPool(ctx,
		pattern.Name("mapper workerpool"),
		pattern.WorkerCount(8),
		pattern.BufferSize(8),
		pattern.Final(),
		pattern.WithWorker(&workers.Mapper{}),
		pattern.WorkerConfigBytes([]byte("$$")),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// to demonstrate we can gracefully start -> process -> finish a complete pipeline
	// concurrently multiple times from start to finish without threading issues
	for i := 0; i < 3; i++ {
		inputBytes := workers.SendEvents(5000)

		// a tree of instances that form a pipeline
		workerPipeline := pattern.PoolTree{
			// mappingWorkerPool worker takes in http request bodies and transforms them
			Instance: mappingWorkerPool,
			/*
				Children: []pattern.PoolTree{
					{
						Instance: mappingWorkerPool,
					},
				},
			*/
		}

		workerPipeline.Start(inputBytes)
	}
}
