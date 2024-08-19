package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/tbal999/pipelines/examples/readers"
	"github.com/tbal999/pipelines/examples/workers"
	pipelines "github.com/tbal999/pipelines/pkg"
	"gopkg.in/yaml.v2"
)

/*
	let's build a piece of software that:

	- takes in events
	- dedupes each event on ingress
	- logs only new events
*/

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(60 * time.Second)
		cancel()
	}()

	configMap := configYAMLtoMap()

	deduperPool, err := pipelines.NewPool(ctx,
		pipelines.Name("file-deduper"),
		pipelines.WorkerCount(1),
		pipelines.BufferSize(500),
		pipelines.WithWorker(&workers.Deduper{}),
		pipelines.WorkerConfigBytes(mapToYaml(configMap, "deduper")),
		pipelines.ErrorHandler(func(err error) {
			if errors.Is(err, pipelines.ErrInit) {
				log.Println(err)
				cancel()
			} else {
				log.Println(err)
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	loggerPool, err := pipelines.NewPool(ctx,
		pipelines.Name("logger"),
		pipelines.WorkerCount(1),
		pipelines.BufferSize(500),
		pipelines.WithWorker(&workers.Logger{}),
		pipelines.WorkerConfigBytes(mapToYaml(configMap, "logger")),
		pipelines.ErrorHandler(func(err error) {
			if errors.Is(err, pipelines.ErrInit) {
				log.Println(err)
				cancel()
			} else {
				log.Println(err)
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1 ; i <= 20; i++ {
		time.Sleep(1 * time.Second)
		readerChannel := readers.SendEvents(ctx, i)

		workerPipeline := pipelines.PoolTree{
			WorkerPool: deduperPool,
			PublishTo: []pipelines.PoolTree{
				{
					WorkerPool: loggerPool,
				},
			},
		}

		workerPipeline.Start(readerChannel)
	}
}

// grab the config file and use it to configure the microservice
func configYAMLtoMap() map[string]any {
	byts, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	out := make(map[string]interface{})
	err = yaml.Unmarshal(byts, &out)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

// grab keys in the map to configure worker pools
func mapToYaml(input map[string]any, key string) []byte {
	val, ok := input[key]
	if !ok {
		log.Fatal(key, " is missing in the config map")
	}

	byts, err := yaml.Marshal(val)
	if err != nil {
		log.Fatal(err)
	}
	return byts
}
