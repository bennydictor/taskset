package middlewares

import (
	"context"

	"github.com/bennydictor/taskset"
)

// NewRecover creates a middleware for handling panics. If a task
// panics, it will be recovered, and its value be given to handle().
// handle() returns a result that becomes the result of the task.
func NewRecover(handle func(interface{}) taskset.Result) taskset.Middleware {
	return taskset.Middleware{
		Run: func(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) (result taskset.Result) {
			defer func() {
				if err := recover(); err != nil {
					result = handle(err)
				}
			}()

			result = next(ctx)
			return
		},
	}
}
