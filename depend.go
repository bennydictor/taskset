package taskset

import (
	"context"
	"sync"
	"sync/atomic"
)

// Depend is used by RunFunc to declare a dependency on another task from the same task set.
// DO NOT use TaskSet.Result to get a result of a task from the same task set.
// Depend will block until the dependent task's results are ready.
// To wait for multiple tasks in parallel, use Depend.ErrGroup or Depend.SyncGroup.
// Depend will implicitly start lazy tasks if they weren't running already.
type Depend func(context.Context, *Task) Result

// ErrGroup starts waiting for a list of tasks in parallel.
// If any of the tasks fail, ErrGroup won't wait for the other tasks' results
// and will return the failed task immediately.
// In this case, context will NOT be cancelled for the other tasks.
// If all tasks complete successfully, ErrGroup will return nil.
func (depend Depend) ErrGroup(ctx context.Context, tasks ...*Task) (result *Task) {
	doneCount := uint32(0)
	done := make(chan struct{})
	var doneOnce sync.Once

	for _, task := range tasks {
		task := task
		go func() {
			if depend(ctx, task).Err != nil {
				doneOnce.Do(func() {
					result = task
					close(done)
				})
			}
			if atomic.AddUint32(&doneCount, 1) == uint32(len(tasks)) {
				doneOnce.Do(func() { close(done) })
			}
		}()
	}

	<-done
	return
}

// SyncGroup waits for a list of tasks in parallel, blocking until every task has returned a result,
// successful or not.
//
// SyncGroup(ctx, tasks...) is different from calling depend() on each task in succession:
// SyncGroup ensures that all tasks are actually running before waiting for all of them to complete,
// whereas waiting for tasks in succession won't start running a lazy task until every task before it completes.
func (depend Depend) SyncGroup(ctx context.Context, tasks ...*Task) {
	doneCount := uint32(0)
	done := make(chan struct{})

	for _, task := range tasks {
		task := task
		go func() {
			depend(ctx, task)
			if atomic.AddUint32(&doneCount, 1) == uint32(len(tasks)) {
				close(done)
			}
		}()
	}

	<-done
}
