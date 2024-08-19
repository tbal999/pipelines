package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/tbal999/pipelines/examples/readers"
	"github.com/tbal999/pipelines/examples/workers"
	pipelines "github.com/tbal999/pipelines/pkg"
	"github.com/tbal999/pipelines/pkg/mocks"

	"go.uber.org/mock/gomock"
)

func TestRun(t *testing.T) {
	t.Run("full pipeline with 5 second context timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(5 * time.Second)
			cancel()
		}()

		mappingWorkerPool, err := pipelines.NewPool(ctx,
			pipelines.Name("logger workerpool 1"),
			pipelines.WorkerCount(2),
			pipelines.BufferSize(4),
			pipelines.WithWorker(&workers.Logger{Log: true}),
			pipelines.WorkerConfigBytes([]byte(``)),
			pipelines.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		mappingWorkerPool2, err := pipelines.NewPool(ctx,
			pipelines.Name("logger workerpool 2"),
			pipelines.WorkerCount(2),
			pipelines.BufferSize(4),
			pipelines.WithWorker(&workers.Logger{Log: true}),
			pipelines.WorkerConfigBytes([]byte(``)),
			pipelines.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		mappingWorkerPool3, err := pipelines.NewPool(ctx,
			pipelines.Name("logger workerpool 3"),
			pipelines.WorkerCount(2),
			pipelines.BufferSize(4),
			pipelines.Final(),
			pipelines.WithWorker(&workers.Logger{Log: true}),
			pipelines.WorkerConfigBytes([]byte(``)),
			pipelines.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		// a tree of worker pools that form a pipeline
		// 1 sends to 2
		// 2 sends to 3
		workerPipeline := pipelines.PoolTree{
			// Converts the input from Reader into JSON
			WorkerPool: mappingWorkerPool,
			PublishTo: []pipelines.PoolTree{
				{
					// Maps the JSON data using JSONATA
					WorkerPool: mappingWorkerPool2,
					PublishTo: []pipelines.PoolTree{
						{
							// Publishes the mapped data to an egress
							WorkerPool: mappingWorkerPool3,
						},
					},
				},
			},
		}

		// Reader is here
		// cancelling context here cancels everything else
		// through a cascade of closing channels
		inputBytes := readers.SendEvents(ctx, 20)

		workerPipeline.Start(inputBytes)
	})

	t.Run("pipeline with mock worker to assert worker behaviour", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(5 * time.Second)
			cancel()
		}()

		mockWorker := mocks.NewMockWorker(ctrl)

		mockworkerPool, err := pipelines.NewPool(ctx,
			pipelines.Name("mock worker"),
			pipelines.WorkerCount(2),
			pipelines.BufferSize(4),
			pipelines.Final(),
			pipelines.WithWorker(mockWorker),
			pipelines.WorkerConfigBytes([]byte(``)),
			pipelines.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		// a tree of worker pools that form a pipeline
		workerPipeline := pipelines.PoolTree{
			WorkerPool: mockworkerPool,
		}

		mockWorker.EXPECT().Close().Return(nil).Times(2)
		mockWorker.EXPECT().Action(gomock.Any()).Return(nil, true, nil).Times(20)
		mockWorker.EXPECT().Initialise(gomock.Any()).Return(nil).Times(2)
		mockWorker.EXPECT().Clone().Return(mockWorker).Times(2)

		// Reader is here
		// cancelling context here cancels everything else
		// through a cascade of closing channels
		inputBytes := readers.SendEvents(ctx, 20)

		workerPipeline.Start(inputBytes)
	})
}
