package main

import (
	"context"
	"log"
	"sync"
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

	drainWorkerPool(rowLoggingPool2)

	time.Sleep(2 * time.Second)

	workerPipeline.Initialise()

	inputChan2, err := workers.ReadCSVWithHeader("./testdata/example.csv")
	if err != nil {
		log.Fatal(err)
	}

	workerPipeline.Start(inputChan2)

	drainWorkerPool(rowLoggingPool2)
}

func drainWorkerPool(workerPool *pattern.Pool) {
	poolChannels, _ := workerPool.GetOutputChannels(), workerPool.Name() // rowLoggingPool

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for index := range poolChannels {
			wg.Add(1)

			go func() {
				defer wg.Done()
				for range poolChannels[index] {
					// drain the channel
				}
			}()
		}
	}()

	wg.Wait()

	log.Println("Gracefully stopped")
}
