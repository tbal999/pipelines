package main

import (
	"log"
	"sync"
	"time"

	pipeline "github.com/tbal999/pipelines/pattern"
)

func main() {
	// take in json {"number": 1} and multiply the number by 2
	poolA, err := pipeline.NewPool(
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
	poolB, err := pipeline.NewPool(
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
	poolC, err := pipeline.NewPool(
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

	poolD, err := pipeline.NewPool(
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
				Children: []pipeline.PoolMap{},
			},
			{
				// this worker consumes from poolA
				Instance: &poolC,
				// and publishes to these workers
				Children: []pipeline.PoolMap{
					{
						// this worker consumes from poolC
						Instance: &poolD,
					},
				},
			},
		},
	}

	examplePipeline(&pipeline)

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
}

func examplePipeline(pools *pipeline.PoolMap) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	pipeline.Initialise(pools)

	var inputChan = make(chan []byte)

	go func() {
		defer wg.Done()
		// wait 5 seconds before passing in empty data
		time.Sleep(1 * time.Second)
		for i := 0; i < 1; i++ {
			inputChan <- []byte(`{"number": 1}`)
		}
		close(inputChan)
	}()

	pipeline.Start(pools, inputChan)

	wg.Wait()
}
