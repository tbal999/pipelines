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
		time.Sleep(1 * time.Second)
	}()

	rowLoggingPool, err := pattern.NewPool(ctx,
		pattern.Name("csv row logging workerpool"),
		pattern.WorkerCount(100),
		pattern.WithWorker(&workers.RowLogger{}),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	rowLoggingPool2, err := pattern.NewPool(ctx,
		pattern.Name("csv row logging workerpool"),
		pattern.WorkerCount(100),
		pattern.NoOutput(),
		pattern.WithWorker(&workers.RowLogger{}),
		pattern.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// a tree of instances that form a pipeline
	workerPipeline := pattern.PoolTree{
		// rowLoggingPool worker takes in bytes & logs them
		Instance: rowLoggingPool,
		Children: []pattern.PoolTree{
			{
				Instance: rowLoggingPool2,
			},
		},
	}

	workerPipeline.Initialise()

	inputChan, err := workers.ReadCSVWithHeader("./testdata/example.csv")
	if err != nil {
		log.Fatal(err)
	}

	workerPipeline.Start(inputChan)

	time.Sleep(2 * time.Second)

	workerPipeline.Initialise()

	inputChan2, err := workers.ReadCSVWithHeader("./testdata/example.csv")
	if err != nil {
		log.Fatal(err)
	}

	workerPipeline.Start(inputChan2)
}
