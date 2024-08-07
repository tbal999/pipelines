package main

import (
	"encoding/json"
	"errors"
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
		pipeline.InitFunc(func() error {
			log.Println("pool A starting")
			return nil
		}),
		pipeline.CloseFunc(func() error {
			log.Println("pool A stopping")
			return nil
		}),
		pipeline.Function(func(input []byte) ([]byte, error) {
			type Input struct {
				Number int
			}

			var item Input
			_ = json.Unmarshal(input, &item)

			item.Number = item.Number * 2

			return json.Marshal(item)
		}),
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
		pipeline.InitFunc(func() error {
			log.Println("pool B starting")
			return nil
		}),
		pipeline.CloseFunc(func() error {
			log.Println("pool B stopping")
			return nil
		}),
		pipeline.Function(func(input []byte) ([]byte, error) {
			type Input struct {
				Number int
			}

			var item Input
			_ = json.Unmarshal(input, &item)

			item.Number = item.Number * 3

			return json.Marshal(item)
		}),
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
		pipeline.InitFunc(func() error {
			log.Println("pool C starting")
			return nil
		}),
		pipeline.FailFast(),
		pipeline.CloseFunc(func() error {
			log.Println("pool C stopping")
			return nil
		}),
		pipeline.Function(func(input []byte) ([]byte, error) {
			type Input struct {
				Number int
			}

			var item Input
			_ = json.Unmarshal(input, &item)

			item.Number = item.Number * 4

			if item.Number % 2 == 0 {
				return nil, errors.New("argh no")
			}

			bytes, _ := json.Marshal(item)

			return bytes, nil
		}),
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
		pipeline.InitFunc(func() error {
			log.Println("pool D starting")
			return nil
		}),
		pipeline.CloseFunc(func() error {
			log.Println("pool D stopping")
			return nil
		}),
		pipeline.Function(func(input []byte) ([]byte, error) {
			type Input struct {
				Number int
			}

			var item Input
			_ = json.Unmarshal(input, &item)

			item.Number = item.Number * 5

			return json.Marshal(item)
		}),
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
	pipeline.Initialise(pools)

	var inputChan = make(chan []byte)

	go func() {
		// wait 5 seconds before passing in empty data
		time.Sleep(1 * time.Second)
		for i := 0; i < 3; i++ {
			inputChan <- []byte(`{"number": 1}`)
		}
		close(inputChan)
	}()

	pipeline.Start(pools, inputChan)
}
