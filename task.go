package taskset

import (
	"context"
	"reflect"
	"sync"
)

// Task is the basic unit of work and concurrency.
// Each task runs in a separate goroutine. A task may depend
// on other tasks' results.
//
// Tasks are created by a TaskSet using a RunFunc and Properties.
type Task struct {
	taskSet *TaskSet

	propertiesMu sync.Mutex
	properties   map[interface{}]interface{}

	once    sync.Once
	started bool
	done    chan struct{}
	run     RunFunc
	result  Result
}

// Result is the result of running a Task.
// Result is considered a success if Err == nil,
// and a failure if Err != nil.
type Result struct {
	Value interface{}
	Err   error
}

// RunFunc is the body of a Task.
// This function will be run in a separate goroutine.
// Context is passed to this function from TaskSet.Start.
//
// The returned values will be stored in the task's Result.
//
// A task may depend on other tasks' results.
// To declare a dependency, use the provided Depend function,
// DO NOT use TaskSet.Result.
type RunFunc func(context.Context, Depend) (interface{}, error)

func newTask(taskSet *TaskSet, run RunFunc) *Task {
	return &Task{
		taskSet:    taskSet,
		properties: make(map[interface{}]interface{}),
		done:       make(chan struct{}),
		run:        run,
	}
}

func (t *Task) dependFunc(ctx context.Context, dependency *Task) Result {
	if t.taskSet != dependency.taskSet {
		panic("dependency is from a different task set")
	}

	return t.taskSet.middleware.Depend(ctx, t, dependency, func(ctx context.Context) Result {
		return dependency.depend(ctx)
	})
}

func (t *Task) depend(ctx context.Context) Result {
	t.once.Do(func() {
		go func() {
			defer close(t.done)

			t.result = t.taskSet.middleware.Run(ctx, t, func(ctx context.Context) (result Result) {
				result.Value, result.Err = t.run(ctx, t.dependFunc)
				return
			})
		}()
	})

	return t.wait(ctx)
}

func (t *Task) wait(ctx context.Context) Result {
	select {
	case <-ctx.Done():
		return Result{Err: ctx.Err()}
	case <-t.done:
		return t.result
	}
}

// Property retrieves this task's property by the given key.
// If there's no property for the given key, nil is returned.
//
// This method should only be used by Middlewares.
func (t *Task) Property(key interface{}) interface{} {
	if key == nil {
		panic("nil key")
	}

	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	t.propertiesMu.Lock()
	defer t.propertiesMu.Unlock()

	return t.properties[key]
}

// ModifyProperty modifies a property for this task.
// The modify function runs under a mutex, to allow for serializable transactions.
// Because of that, you shouldn't do any long operations in the modify function.
//
// You should define your own key type to use with this method, like for context.WithValue.
func (t *Task) ModifyProperty(key interface{}, modify func(interface{}) interface{}) {
	if key == nil {
		panic("nil key")
	}

	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	t.propertiesMu.Lock()
	defer t.propertiesMu.Unlock()

	t.properties[key] = modify(t.properties[key])
}
