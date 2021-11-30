package taskset

import (
	"context"
)

// TaskSet creates and runs Tasks.
type TaskSet struct {
	middleware Middleware

	eagerTasks []*Task
}

// NewTaskSet creates a new TaskSet.
func NewTaskSet(middlewares ...Middleware) *TaskSet {
	mw := chainMiddlewares(middlewares)
	if mw.Run == nil {
		mw.Run = func(ctx context.Context, _ *Task, next func(ctx context.Context) Result) Result { return next(ctx) }
	}
	if mw.Depend == nil {
		mw.Depend = func(ctx context.Context, _, _ *Task, next func(ctx context.Context) Result) Result { return next(ctx) }
	}
	return &TaskSet{middleware: mw}
}

// New creates a new Task given its RunFunc and Properties.
// The created task will run upon calling Start.
func (ts *TaskSet) New(run RunFunc, properties ...Property) *Task {
	task := ts.NewLazy(run, properties...)
	ts.eagerTasks = append(ts.eagerTasks, task)
	return task
}

// NewLazy is like New, except the created Task will not run automatically upon
// calling Start.
//
// The created Task will start running the first time any other task declares a
// dependency on it. If no tasks declare dependency on the created task, it will
// not run at all.
//
// A lazy task can be later converted to a non-lazy with Eager.
func (ts *TaskSet) NewLazy(run RunFunc, properties ...Property) *Task {
	task := newTask(ts, run)
	for _, p := range properties {
		p(task)
	}
	return task
}

// Eager marks a lazy task to be non-lazy. The provided task must belong to this TaskSet.
func (ts *TaskSet) Eager(task *Task) {
	if task.taskSet != ts {
		panic("task doesn't belong to task set")
	}

	ts.eagerTasks = append(ts.eagerTasks, task)
}

// Start runs all non-lazy Tasks created by this task set.
// Context will be passed to all the tasks' run functions.
func (ts *TaskSet) Start(ctx context.Context) {
	ctx = context.WithValue(ctx, ts, struct{}{})

	for _, task := range ts.eagerTasks {
		task := task
		go func() { _ = task.depend(ctx) }()
	}
}

// Wait waits for all non-lazy tasks to complete.
// Context is only used to cancel Wait, it is not passed to any of the tasks' RunFuncs.
//
// Wait does not run any tasks, it only waits for them to finish. If you call Wait
// and never call Start, it will block forever.
func (ts *TaskSet) Wait(ctx context.Context) {
	for _, task := range ts.eagerTasks {
		task.wait(ctx)
	}
}

// Result returns the Result of a given Task, blocking until it is ready.
// The provided task must belong to this TaskSet.
// Context is only used to cancel Result, it is not passed to any of the tasks' RunFuncs.
//
// Result does not run a task, it only waits for the result. If you call Result
// on a task that is never run, it will block forever.
//
// DO NOT use Result from inside a RunFunc to get a result of a task from the same task set.
// Such a call to Result will panic. You should use depend(ctx, task) instead.
func (ts *TaskSet) Result(ctx context.Context, task *Task) Result {
	if task.taskSet != ts {
		panic("task doesn't belong to task set")
	}

	if ctx.Value(ts) != nil {
		panic("Don't call taskSet.Result(task) from inside tasks. Use depend(task) instead.")
	}

	return task.wait(ctx)
}
