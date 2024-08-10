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
	ctx := context.Background()

	rowLoggingPool, err := pattern.NewPool(ctx,
		pattern.Name("csv row logging workerpool"),
		pattern.WorkerCount(1),
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
		pattern.WorkerCount(1),
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

	pattern.Initialise(&workerPipeline)

	inputChan, err := workers.ReadCSVWithHeader("./testdata/example.csv")
	if err != nil {
		log.Fatal(err)
	}

	pattern.Start(&workerPipeline, inputChan)

	drain(rowLoggingPool2)

	time.Sleep(2 * time.Second)

	pattern.Initialise(&workerPipeline)

	inputChan2, err := workers.ReadCSVWithHeader("./testdata/example.csv")
	if err != nil {
		log.Fatal(err)
	}

	pattern.Start(&workerPipeline, inputChan2)

	drain(rowLoggingPool2)
}

func drain(workerPool *pattern.Pool) {
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

	log.Println("Waiting gracefully...")

	wg.Wait()

	log.Println("Gracefully stopped")
}
