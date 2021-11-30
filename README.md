# taskset
[![Go Reference](https://pkg.go.dev/badge/github.com/bennydictor/taskset.svg)](https://pkg.go.dev/github.com/bennydictor/taskset)

Taskset is a library for running concurrent tasks.
Tasks can do arbitrary work and depend on other tasks' results.

You can decorate the behaviour of all your tasks with middlewares, for example,
you can add logging, measure execution duration, or limit the number of tasks allowed to run at the same time.
Here's a [list of standard middlewares](https://pkg.go.dev/github.com/bennydictor/taskset/middlewares).

A basic example:
```go
ctx := context.Background()

taskSet := taskset.NewTaskSet()

taskA := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
	time.Sleep(2 * time.Second)
	return 1, nil
})

taskB := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
	time.Sleep(2 * time.Second)
	return 2, nil
})

taskC := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
	if t := depend.ErrGroup(ctx, taskA, taskB); t != nil {
		return nil, depend(ctx, t).Err
	}

	a := depend(ctx, taskA).Value.(int)
	b := depend(ctx, taskB).Value.(int)

	time.Sleep(2 * time.Second)
	return a + b, nil
})

start := time.Now()
taskSet.Start(ctx)
taskSet.Wait(ctx)
totalTime := time.Since(start)

fmt.Printf("total time: %.0fs\n", totalTime.Seconds())
fmt.Println("result:", taskSet.Result(ctx, taskC).Value)

// Output:
// total time: 4s
// result: 3
```

## 3rd-party middlewares

We also provide some middlewares that integrate with 3rd-party libraries:

- [zap](https://github.com/uber-go/zap) logging framework: [![Go Reference](https://pkg.go.dev/badge/github.com/bennydictor/taskset/middlewares/zap.svg)](https://pkg.go.dev/github.com/bennydictor/taskset/middlewares/zap)
- [prometheus](https://prometheus.io/) metrics collection: [![Go Reference](https://pkg.go.dev/badge/github.com/bennydictor/taskset/middlewares/prometheus.svg)](https://pkg.go.dev/github.com/bennydictor/taskset/middlewares/prometheus)
- [opentracing](https://opentracing.io/) API for tracing: [![Go Reference](https://pkg.go.dev/badge/github.com/bennydictor/taskset/middlewares/opentracing.svg)](https://pkg.go.dev/github.com/bennydictor/taskset/middlewares/opentracing)
