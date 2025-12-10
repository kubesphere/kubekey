package utils

import (
	"context"
	"sync"
)

type Worker[T any] struct {
	MaxWorkerCount int
	TaskChan       chan T
	ExecFunc       func(T) error
	wg             sync.WaitGroup
	results        chan error
}

func (worker *Worker[T]) Do(ctx context.Context) {
	worker.results = make(chan error)
	for w := 0; w < worker.MaxWorkerCount; w++ {
		worker.wg.Add(1)
		go func(workerID int) {
			defer worker.wg.Done()
			for thing := range worker.TaskChan {
				select {
				case <-ctx.Done():
					worker.results <- ctx.Err()
					return
				default:
					worker.results <- worker.ExecFunc(thing)
				}
			}
		}(w)
	}
}

func (worker *Worker[T]) Wait() {
	worker.wg.Wait()
	close(worker.results)
}

func (worker *Worker[T]) CollectedErrors() []string {
	var collectedErrors []string
	for err := range worker.results {
		if err != nil {
			collectedErrors = append(collectedErrors, err.Error())
		}
	}
	return collectedErrors
}
