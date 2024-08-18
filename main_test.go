package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/tbal999/pipelines/pattern"
	"github.com/tbal999/pipelines/pattern/mocks"
	"github.com/tbal999/pipelines/readers"
	"github.com/tbal999/pipelines/workers"

	"go.uber.org/mock/gomock"
)

func TestRun(t *testing.T) {
	t.Run("full pipeline with 5 second context timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(5 * time.Second)
			cancel()
		}()

		mappingWorkerPool, err := pattern.NewPool(ctx,
			pattern.Name("mapper workerpool 1"),
			pattern.WorkerCount(2),
			pattern.BufferSize(4),
			pattern.WithWorker(&workers.Mapper{Log: true}),
			pattern.WorkerConfigBytes([]byte(``)),
			pattern.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		mappingWorkerPool2, err := pattern.NewPool(ctx,
			pattern.Name("mapper workerpool 2"),
			pattern.WorkerCount(2),
			pattern.BufferSize(4),
			pattern.WithWorker(&workers.Mapper{Log: true}),
			pattern.WorkerConfigBytes([]byte(``)),
			pattern.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		mappingWorkerPool3, err := pattern.NewPool(ctx,
			pattern.Name("mapper workerpool 3"),
			pattern.WorkerCount(2),
			pattern.BufferSize(4),
			pattern.Final(),
			pattern.WithWorker(&workers.Mapper{Log: true}),
			pattern.WorkerConfigBytes([]byte(``)),
			pattern.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		// a tree of worker pools that form a pipeline
		// 1 sends to 2
		// 2 sends to 3
		workerPipeline := pattern.PoolTree{
			// Converts the input from Reader into JSON
			WorkerPool: mappingWorkerPool,
			PublishTo: []pattern.PoolTree{
				{
					// Maps the JSON data using JSONATA
					WorkerPool: mappingWorkerPool2,
					PublishTo: []pattern.PoolTree{
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

		mockworkerPool, err := pattern.NewPool(ctx,
			pattern.Name("mock worker"),
			pattern.WorkerCount(2),
			pattern.BufferSize(4),
			pattern.Final(),
			pattern.WithWorker(mockWorker),
			pattern.WorkerConfigBytes([]byte(``)),
			pattern.ErrorHandler(func(err error) {
				log.Println(err.Error())
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		// a tree of worker pools that form a pipeline
		workerPipeline := pattern.PoolTree{
			WorkerPool: mockworkerPool,
		}

		mockWorker.EXPECT().Close().Return(nil).Times(2)
		mockWorker.EXPECT().Action(gomock.Any()).Return(nil, nil).Times(20)
		mockWorker.EXPECT().Initialise(gomock.Any()).Return(nil).Times(2)
		mockWorker.EXPECT().Clone().Return(mockWorker).Times(2)

		// Reader is here
		// cancelling context here cancels everything else
		// through a cascade of closing channels
		inputBytes := readers.SendEvents(ctx, 20)

		workerPipeline.Start(inputBytes)
	})
}
