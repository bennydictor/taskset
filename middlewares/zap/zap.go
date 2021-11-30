package zap

import (
	"context"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/properties"
	"go.uber.org/zap"
)

type disableLoggingProperty struct{}

// WithDisableLogging disables logging for a particular task.
var WithDisableLogging taskset.Property = func(task *taskset.Task) {
	task.ModifyProperty(disableLoggingProperty{}, func(_ interface{}) interface{} {
		return struct{}{}
	})
}

// NewLogger creates a logging middleware. It will log an info message
// when a task is started and successfully finished, and error message when a task is failed,
// and a debug message for every depend() call.
//
// Logging can be disabled for a particular task using WithDisableLogging.
func NewLogger(logger *zap.Logger) taskset.Middleware {
	return taskset.Middleware{
		Run: func(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			if task.Property(disableLoggingProperty{}) != nil {
				return next(ctx)
			}

			log := logger.With(zap.String("task", properties.Name(task)))
			log.Info("task started")

			result := next(ctx)

			if result.Err != nil {
				log.Error("task failed", zap.Error(result.Err))
			} else {
				log.Info("task done")
			}

			return result
		},
		Depend: func(ctx context.Context, task, dependency *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			if task.Property(disableLoggingProperty{}) != nil {
				return next(ctx)
			}

			log := logger.With(
				zap.String("task", properties.Name(task)),
				zap.String("dependency", properties.Name(dependency)),
			)

			log.Debug("depend start")
			result := next(ctx)
			log.Debug("depend end")
			return result
		},
	}
}
