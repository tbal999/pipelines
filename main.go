package main

import (
	"context"
	"log"
	"sync"

	pipeline "github.com/tbal999/pipelines/pattern"
)

func main() {
	ctx := context.Background()

	// take in json {"number": 1} and multiply the number by 2
	poolA, err := pipeline.NewPool(ctx,
		pipeline.Name("pool A"),
		pipeline.WorkerCount(1),
		pipeline.WithWorker(&NumberWorker{Multiplier: 1}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// take in json {"number": 1} and multiply the number by 3
	poolB, err := pipeline.NewPool(ctx,
		pipeline.Name("pool B"),
		pipeline.WorkerCount(2),
		pipeline.WithWorker(&NumberWorker{Multiplier: 7}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// take in json {"number": 1} and multiply the number by 4
	poolC, err := pipeline.NewPool(ctx,
		pipeline.Name("pool C"),
		pipeline.WorkerCount(2),
		pipeline.FailFast(),
		pipeline.WithWorker(&NumberWorker{Multiplier: 2}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	poolD, err := pipeline.NewPool(ctx,
		pipeline.Name("pool D"),
		pipeline.WorkerCount(2),
		pipeline.WithWorker(&NumberWorker{Multiplier: 3}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// a tree of instances that form a pipeline
	workerPipeline := pipeline.PoolTree{
		// poolA worker takes in bytes
		Instance: poolA,
		// and publishes to these workers (poolB and poolC)
		Children: []pipeline.PoolTree{
			{
				// poolB worker consumes from poolA
				Instance: poolB,
				// no children so this has 1 output channel which _needs_ to be read from
			},
			{
				// poolC worker consumes from poolA
				Instance: poolC,
				// and publishes to these workers (poolD)
				Children: []pipeline.PoolTree{
					{
						// this worker consumes from poolC
						Instance: poolD,
						// no children so this has 1 output channel which _needs_ to be read from
					},
				},
			},
		},
	}

	pipeline.Initialise(&workerPipeline)

	poolBOutput, poolBName := poolB.GetOutputChannels(), poolB.Name() // poolB
	poolDOutput, poolDName := poolD.GetOutputChannels(), poolD.Name() // poolD

	var inputChan = make(chan []byte)

	go func() {
		defer func() {
			close(inputChan)
			log.Println("closed input channel")
		}()

		for i := 0; i < 4; i++ {
			inputChan <- []byte(`{"number": 1}`)
		}
	}()

	pipeline.Start(&workerPipeline, inputChan)

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		for index := range poolBOutput {
			wg.Add(1)

			go func() {
				defer wg.Done()
				for item := range poolBOutput[index] {
					log.Println(poolBName, string(item))
				}
			}()
		}
	}()

	go func() {
		defer wg.Done()

		for index := range poolDOutput {
			wg.Add(1)

			go func() {
				defer wg.Done()
				for item := range poolDOutput[index] {
					log.Println(poolDName, string(item))
				}
			}()
		}
	}()

	log.Println("Waiting gracefully...")

	wg.Wait()

	log.Println("Gracefully stopped")
}
