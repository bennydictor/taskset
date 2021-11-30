package taskset_test

import (
	"context"
	"fmt"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/properties"
)

func NewPrinter() taskset.Middleware {
	return taskset.Middleware{
		Run: func(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			fmt.Println(properties.Name(task), "starting")
			result := next(ctx)
			fmt.Println(properties.Name(task), "finished")
			return result
		},
		Depend: func(ctx context.Context, task, dependency *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			fmt.Println(properties.Name(task), "depend on", properties.Name(dependency), "starting")
			result := next(ctx)
			fmt.Println(properties.Name(task), "depend on", properties.Name(dependency), "finished")
			return result
		},
	}
}

func ExampleMiddleware() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet(
		NewPrinter(),
	)

	taskA := taskSet.NewLazy(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		return 1, nil
	},
		properties.WithName("A"),
	)

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		a := depend(ctx, taskA).Value.(int)
		return a + 1, nil
	},
		properties.WithName("B"),
	)

	taskSet.Start(ctx)
	taskSet.Wait(ctx)

	// Output:
	// B starting
	// B depend on A starting
	// A starting
	// A finished
	// B depend on A finished
	// B finished
}
