package main

import (
	"context"
	"testing"

	pipelines "github.com/tbal999/pipelines/pkg"
	"github.com/tbal999/pipelines/readers"
	"github.com/tbal999/pipelines/workers"
)

func BenchmarkPipeline(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool1, _ := pipelines.NewPool(ctx, pipelines.Name("Stage 1"), pipelines.WorkerCount(2), pipelines.BufferSize(4), pipelines.WithWorker(&workers.Mapper{Log: false}))
	pool2, _ := pipelines.NewPool(ctx, pipelines.Name("Stage 2"), pipelines.WorkerCount(2), pipelines.BufferSize(4), pipelines.WithWorker(&workers.Mapper{Log: false}))
	pool3, _ := pipelines.NewPool(ctx, pipelines.Name("Stage 3"), pipelines.WorkerCount(2), pipelines.BufferSize(4), pipelines.WithWorker(&workers.Mapper{Log: false}), pipelines.Final())

	pipeline := pipelines.PoolTree{
		WorkerPool: pool1,
		PublishTo: []pipelines.PoolTree{
			{
				WorkerPool: pool2,
				PublishTo: []pipelines.PoolTree{
					{
						WorkerPool: pool3,
					},
				},
			},
		},
	}

	for n := 0; n < b.N; n++ {
		iterCtx, iterCancel := context.WithCancel(context.Background())
		defer iterCancel()

		inputData := readers.SendEvents(iterCtx, 500000)
		pipeline.Start(inputData)
	}
}

/* 500000 events in 0.425 seconds */ 