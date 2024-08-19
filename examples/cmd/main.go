package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/tbal999/pipelines/examples/readers"
	"github.com/tbal999/pipelines/examples/workers"
	pipelines "github.com/tbal999/pipelines/pkg"
	"gopkg.in/yaml.v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Second)
		cancel()
	}()

	configMap := yamlToMap()

	/*
		let's build a piece of software that:

		- takes in HTTP events
		-> routes those events to a logger
		-> routes the events to a local file writer
		-> any errors are captured and routed back to the main application
		-> the entire application gracefully shuts down with a context cancel
	*/

	loggerPool, err := pipelines.NewPool(ctx,
		pipelines.Name("logger"),
		pipelines.WorkerCount(2),
		pipelines.BufferSize(4),
		pipelines.WithWorker(&workers.Logger{}),
		pipelines.WorkerConfigBytes(mapToYaml(configMap["logger"])),
		pipelines.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fileWriter, err := pipelines.NewPool(ctx,
		pipelines.Name("file-deduper"),
		pipelines.WorkerCount(2),
		pipelines.BufferSize(4),
		pipelines.WithWorker(&workers.Deduper{}),
		pipelines.WorkerConfigBytes(mapToYaml(configMap["deduper"])),
		pipelines.ErrorHandler(func(err error) {
			log.Println(err.Error())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	readerChannel := readers.HTTPEvents(ctx)

	workerPipeline := pipelines.PoolTree{
		WorkerPool: fileWriter,
		PublishTo: []pipelines.PoolTree{
			{
				WorkerPool: loggerPool,
			},
		},
	}

	workerPipeline.Start(readerChannel)
}

func yamlToMap() map[string]any {
	byts, _ := os.ReadFile("config.yaml")
	out := make(map[string]interface{})
	_ = yaml.Unmarshal(byts, &out)
	return out
}

func mapToYaml(input interface{}) []byte {
	byts, _ := yaml.Marshal(input)
	return byts
}
