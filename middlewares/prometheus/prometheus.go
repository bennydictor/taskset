package prometheus

import (
	"context"
	"time"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/properties"
	"github.com/prometheus/client_golang/prometheus"
)

type disableMetricsProperty struct{}

// WithDisableMetrics disables metric collection for a particular task.
var WithDisableMetrics taskset.Property = func(task *taskset.Task) {
	task.ModifyProperty(disableMetricsProperty{}, func(_ interface{}) interface{} {
		return struct{}{}
	})
}

// Metrics is a set of metrics reported by prometheus middleware.
type Metrics struct {
	// Duration is used to report each task's run duration in seconds.
	// The reported duration will exclude the time a task spends in depend() calls.
	Duration *prometheus.HistogramVec
	// Success is used to report each task's successful execution.
	Success *prometheus.CounterVec
	// Failure is used to report each task's failed execution.
	Failure *prometheus.CounterVec
}

type timerProperty struct{}

type timer struct {
	dependCount uint
	total       time.Duration
	start       time.Time
}

// NewPrometheus creates a middleware that reports basic metrics on each task.
//
// Metric collection can be disabled for a particular task using WithDisableMetrics.
//
// Every metric is reported using label values equal to {properties.Name(task)}.
// If any of the metrics are nil, then that metric won't be reported.
func NewPrometheus(metrics Metrics) taskset.Middleware {
	return taskset.Middleware{
		Run: func(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			if task.Property(disableMetricsProperty{}) != nil {
				return next(ctx)
			}

			if metrics.Duration != nil {
				task.ModifyProperty(timerProperty{}, func(_ interface{}) interface{} {
					return timer{
						start: time.Now(),
					}
				})
			}

			result := next(ctx)

			if metrics.Duration != nil {
				t := task.Property(timerProperty{}).(timer)
				duration := t.total + time.Since(t.start)
				observer, err := metrics.Duration.GetMetricWithLabelValues(properties.Name(task))
				if err == nil {
					observer.Observe(duration.Seconds())
				}
			}

			if result.Err == nil && metrics.Success != nil {
				counter, err := metrics.Success.GetMetricWithLabelValues(properties.Name(task))
				if err == nil {
					counter.Inc()
				}
			}

			if result.Err != nil && metrics.Failure != nil {
				counter, err := metrics.Failure.GetMetricWithLabelValues(properties.Name(task))
				if err == nil {
					counter.Inc()
				}
			}

			return result
		},
		Depend: func(ctx context.Context, task, dependency *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			if task.Property(disableMetricsProperty{}) != nil {
				return next(ctx)
			}

			if metrics.Duration != nil {
				task.ModifyProperty(timerProperty{}, func(value interface{}) interface{} {
					t := value.(timer)
					if t.dependCount == 0 {
						t.total += time.Since(t.start)
					}
					t.dependCount += 1
					return t
				})
			}

			result := next(ctx)

			if metrics.Duration != nil {
				task.ModifyProperty(timerProperty{}, func(value interface{}) interface{} {
					t := value.(timer)
					t.dependCount -= 1
					if t.dependCount == 0 {
						t.start = time.Now()
					}
					return t
				})
			}

			return result
		},
	}
}
