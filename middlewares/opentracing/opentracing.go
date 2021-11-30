package opentracing

import (
	"context"
	"fmt"
	"time"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/properties"
	"github.com/opentracing/opentracing-go"
)

type disableTracingProperty struct{}

// WithDisableTracing disables tracing for a particular task.
var WithDisableTracing taskset.Property = func(task *taskset.Task) {
	task.ModifyProperty(disableTracingProperty{}, func(_ interface{}) interface{} {
		return struct{}{}
	})
}

type spanOpNameProperty struct{}

// WithSpanOpName sets the span operation name for a task.
func WithSpanOpName(name string) taskset.Property {
	return func(task *taskset.Task) {
		task.ModifyProperty(spanOpNameProperty{}, func(_ interface{}) interface{} {
			return name
		})
	}
}

// StartSpan determines when to start the opentracing span.
type StartSpan int

const (
	// StartSpanOnTaskStart will start the span as soon as the task starts running.
	StartSpanOnTaskStart = StartSpan(iota)
	// StartSpanOnLastDepend will start the as soon as the task's last depend() call is done.
	// This mode is best used when all your tasks resolve all their dependencies as soon
	// as they're started.
	StartSpanOnLastDepend
)

type startTimeProperty struct{}

// NewOpentracing creates a middleware that creates an opentracing span for each task.
//
// Tracing can be disabled for a particular task using WithDisableTracing.
//
// The operation name of the spans can be set using WithSpanOpName. If no name is set,
// it's formed as "taskset:" + properties.Name(task). If a task has no name, the operation
// name will be "taskset:task".
//
// Start time of each span can be controlled using StartSpan.
func NewOpentracing(startSpan StartSpan) taskset.Middleware {
	return taskset.Middleware{
		Run: func(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			if task.Property(disableTracingProperty{}) != nil {
				return next(ctx)
			}

			task.ModifyProperty(startTimeProperty{}, func(_ interface{}) interface{} {
				return time.Now()
			})

			result := next(ctx)

			startTime := task.Property(startTimeProperty{}).(time.Time)

			opName, ok := task.Property(spanOpNameProperty{}).(string)
			if !ok {
				taskName := properties.Name(task)
				if taskName != "" {
					opName = fmt.Sprintf("taskset:%s", taskName)
				} else {
					opName = "taskset:task"
				}
			}

			span, _ := opentracing.StartSpanFromContext(ctx, opName, opentracing.StartTime(startTime))
			span.Finish()

			return result
		},
		Depend: func(ctx context.Context, task, dependency *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			if task.Property(disableTracingProperty{}) != nil {
				return next(ctx)
			}

			result := next(ctx)

			if startSpan == StartSpanOnLastDepend {
				task.ModifyProperty(startTimeProperty{}, func(_ interface{}) interface{} {
					return time.Now()
				})
			}

			return result
		},
	}
}
