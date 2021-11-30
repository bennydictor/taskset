package middlewares

import (
	"context"
	"fmt"
	"log"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/properties"
)

// Logger is a basic logging middleware. It will log using log.Println
// when each task is started and finished.
var Logger = taskset.Middleware{
	Run: run,
}

func taskName(task *taskset.Task) string {
	name := properties.Name(task)
	if name == "" {
		return "an unnamed task"
	} else {
		return fmt.Sprintf("task %s", name)
	}
}

func run(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
	log.Println(taskName(task), "starting")

	result := next(ctx)

	if result.Err != nil {
		log.Println(taskName(task), "failed:", result.Err.Error())
	} else {
		log.Println(taskName(task), "finished successfully")
	}

	return result
}
