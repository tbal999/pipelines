package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	pipeline "github.com/tbal999/pipelines/pkg"
)

func main() {
	poolA, err := pipeline.NewPool(
		pipeline.Name("pool A"),
		pipeline.WorkerCount(5),
		pipeline.Function(func(input []byte) ([]byte, error) {
			return input, nil
		}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	poolB, err := pipeline.NewPool(
		pipeline.Name("pool B"),
		pipeline.WorkerCount(5),
		pipeline.Function(func(input []byte) ([]byte, error) {
			return input, nil
		}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	poolC, err := pipeline.NewPool(
		pipeline.Name("pool C"),
		pipeline.WorkerCount(5),
		pipeline.Function(func(input []byte) ([]byte, error) {
			return input, nil
		}),
		pipeline.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// an array of worker pools that form a pipeline
	pipeline := []pipeline.Pool{
		poolA,
		poolB,
		poolC,
	}

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		prevs := make([]uint64, len(pipeline))
		for index := range prevs {
			prevs[index] = 0
		}

		for range ticker.C {
			output := []string{}
			for index := range pipeline {
				count := pipeline[index].Total()
				totalC := count - prevs[index]
				output = append(output, fmt.Sprintf("%s total processed: %d - rate per second: %d", pipeline[index].Name(), count, totalC))
				prevs[index] = count
			}

			log.Println(strings.Join(output, " | "))
		}
	}()

	examplePipeline(pipeline)

	// here we run it again to demonstrate the graceful startup and shutdown

	examplePipeline(pipeline)

	time.Sleep(1 * time.Second)
}

func examplePipeline(pools []pipeline.Pool) {
	var inputChan = make(chan []byte)

	go func() {
		// wait 5 seconds before passing in empty data
		time.Sleep(1 * time.Second)
		for i := 0; i < 1000000; i++ {
			inputChan <- []byte("")
		}
		close(inputChan)
	}()

	outChan := pipeline.Start(pools, inputChan)

	for range outChan {
		// items are spat out here
	}

	// gracefully close when finished
}
