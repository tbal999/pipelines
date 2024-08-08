package main

import (
	"context"
	"log"
	"sync"
	"time"

	pipeline "github.com/tbal999/pipelines/pattern"
)

func main() {
	// Create a context with a timeout
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	}()

	// take in json {"number": 1} and multiply the number by 2
	poolA, err := pipeline.NewPool(ctx,
		pipeline.Name("pool A"),
		pipeline.WorkerCount(5),
		pipeline.WithWorker(&NumberWorker{Multiplier: 2}, true),
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
		pipeline.WorkerCount(5),
		pipeline.WithWorker(&NumberWorker{Multiplier: 7}, true),
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
		pipeline.WorkerCount(5),
		pipeline.FailFast(),
		pipeline.WithWorker(&NumberWorker{Multiplier: 1}, true),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	poolD, err := pipeline.NewPool(ctx,
		pipeline.Name("pool D"),
		pipeline.WorkerCount(5),
		pipeline.WithWorker(&NumberWorker{Multiplier: 2}, true),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// a tree of instances that form a pipeline
	pipeline := pipeline.PoolMap{
		// this worker takes in bytes
		Instance: &poolA,
		// and publishes to these workers
		Children: []pipeline.PoolMap{
			{
				// this worker consumes from poolA
				Instance: &poolB,
				// no children so this is an output -->>
			},
			{
				// this worker consumes from poolA
				Instance: &poolC,
				// and publishes to these workers
				Children: []pipeline.PoolMap{
					{
						// this worker consumes from poolC
						Instance: &poolD,
						// no children so this is an output -->>
					},
				},
			},
		},
	}

	examplePipeline(ctx, &pipeline)

	// TODO make below easier to work with

	channels1 := pipeline.Children[0].GetChannels()             // poolB
	channels2 := pipeline.Children[1].GetChannels()             // poolC
	channels3 := pipeline.Children[1].Children[0].GetChannels() // poolD

	wg := sync.WaitGroup{}

	for index := range channels1 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for item := range channels1[index] {
				log.Println(pipeline.Children[0].Instance.Name(), string(item))
			}
		}()
	}

	for index := range channels2 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for item := range channels2[index] {
				log.Println(pipeline.Children[1].Instance.Name(), string(item))
			}
		}()
	}

	for index := range channels3 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for item := range channels3[index] {
				log.Println(pipeline.Children[1].Children[0].Instance.Name(), string(item))
			}
		}()
	}

	wg.Wait()

	log.Println("Gracefully stopped")
}

func examplePipeline(ctx context.Context, pools *pipeline.PoolMap) {
	pipeline.Initialise(pools)

	var inputChan = make(chan []byte)

	go func() {
		defer close(inputChan)
		time.Sleep(1 * time.Second)
		// wait 5 seconds before passing in empty data
		for i := 0; i < 100000; i++ {
			select {
			case <-ctx.Done():
				log.Println("CANCELLED----------------------------------------")
				return
			case inputChan <- []byte(`{"number": 1}`):
			}
		}
	}()

	pipeline.Start(pools, inputChan)
}
