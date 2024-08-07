# Pipelines

The pipelines package is a Go module designed for creating and managing concurrent processing pipelines. The package aims to facilitate the execution of high-throughput tasks by utilizing a pool of worker goroutines.

It gracefully stops when the input channel is closed & can be restarted with a new channel. 
This is demonstrated via the main.go package.

Key Features
    Pipeline Creation: The package allows users to define multiple pipelines, each with its own pool of workers.
    Concurrency Management: It leverages goroutines for concurrent task processing, enabling efficient utilization of system resources.
    Error Handling: Custom error handlers can be defined for each pipeline step to manage and log errors effectively.
    Performance Metrics: The package provides mechanisms to monitor and log the performance of each pipeline, including the number of tasks processed and the rate of processing.

Core Components
    Pool: Represents a pool of workers that process tasks concurrently.
    Options: Configurable options for each pool, such as the number of workers, the function to be executed by each worker, and the error handler.
    Pipeline Functions:
        NewPool(opts ...PoolOption) (Pool, error): Creates a new pool with the specified options.
        Start(pools []Pool, inputChan chan []byte) chan []byte: Starts the processing pipeline with the given pools and input channel.
        It then returns an output channel - which is the final channel of the pipeline.