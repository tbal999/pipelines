package main

import (
	"encoding/json"
	"log"

	pipeline "github.com/tbal999/pipelines/pattern"
)

/*
type Worker interface {
	Action(input []byte) ([]byte, error)
	Initialise() error
	Close() error
}
*/

type NumberWorker struct {
	Multiplier int

	internalValue int
}

func (n *NumberWorker) Clone() pipeline.Worker {
	clone := *n

	return &clone
}

func (n *NumberWorker) Initialise() error {
	log.Println("starting worker", n)
	return nil
}

func (n *NumberWorker) Close() error {
	log.Println("stopping worker", n)
	return nil
}

func (n *NumberWorker) Action(input []byte) ([]byte, error) {
	// to test thread safety without mutexes
	n.internalValue = n.internalValue + 1

	type Input struct {
		Number int
	}

	var item Input

	_ = json.Unmarshal(input, &item)

	item.Number = item.Number * n.Multiplier

	return json.Marshal(item)
}
