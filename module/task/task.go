package task

import (
	"context"
	"go-api-boilerplate/module/logger"
	"time"
)

type TaskResult struct {
	RequestData interface{}
	Data        interface{}
	Err         error
}

func HandleTask(ctx context.Context, ch chan TaskResult, fn func() TaskResult) {
	go func() {
		select {
		case ch <- fn():
		case <-ctx.Done():
		}
	}()
}

func AllSuccessOrAnyFail(ch <-chan TaskResult, cancel context.CancelFunc, tasks int) ([]TaskResult, error) {
	var results []TaskResult

	for range tasks {
		result := <-ch
		if result.Err != nil {
			// If any of the the goroutine failed
			// cancel the context to mark goroutines as done
			// return result to break the loop
			cancel()
			return nil, result.Err
		} else {
			results = append(results, result)
		}
	}

	// Checked result from all the goroutines
	return results, nil
}

// task.ExecutionTime returns a function that prints the name argument and
// the elapsed time between the call to module.ExecutionTime and the call to
// the returned function. The returned function is intended to
// be used in a defer statement:
//
//	defer module.ExecutionTime("funcName", time.now())
func ExecutionTime(name string, start time.Time) {
	elapsed := time.Since(start)

	logger.Log.Infof("[%s] took %s", name, elapsed)
}
