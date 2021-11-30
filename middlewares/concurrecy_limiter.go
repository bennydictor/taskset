package middlewares

import (
	"context"
	"sync"

	"github.com/bennydictor/taskset"
	"golang.org/x/sync/semaphore"
)

type semaphoreLocker struct {
	*semaphore.Weighted
}

// NewSemaphore returns a sync.Locker that can be locked up to n times concurrently.
func NewSemaphore(n int64) sync.Locker {
	return semaphoreLocker{semaphore.NewWeighted(n)}
}

// Lock implements sync.Locker.
func (s semaphoreLocker) Lock() {
	_ = s.Acquire(context.Background(), 1)
}

// Unlock implements sync.Locker.
func (s semaphoreLocker) Unlock() {
	s.Release(1)
}

type noopLocker struct{}

func (n noopLocker) Lock() {}

func (n noopLocker) Unlock() {}

type lockProperty struct{}

// WithLock adds a lock to a task to be used with concurrency limiter.
func WithLock(lock sync.Locker) taskset.Property {
	return func(task *taskset.Task) {
		task.ModifyProperty(lockProperty{}, func(_ interface{}) interface{} {
			return lock
		})
	}
}

type dependCountProperty struct{}

// NewConcurrencyLimiter creates a middleware used to limit a task set's concurrency.
// Concurrency limiter will lock and unlock the provided lock before and after
// each task is run, releasing the lock while the task is in the process of depending
// on another task.
//
// If lock is nil, concurrency limiter will instead use a lock provided
// for each task separately using WithLock. If no lock was provided for a task,
// nothing is locked for that task.
//
// If you want to run all tasks sequentially, use &sync.Mutex{}.
// If you want to limit the number of parallel tasks, use NewSemaphore.
// If you want a subset of tasks to be mutually exclusive, use WithLock.
func NewConcurrencyLimiter(lock sync.Locker) taskset.Middleware {
	getLock := func(task *taskset.Task) sync.Locker {
		if lock != nil {
			return lock
		}

		if taskLock := task.Property(lockProperty{}); taskLock != nil {
			return taskLock.(sync.Locker)
		}

		return noopLocker{}
	}

	return taskset.Middleware{
		Run: func(ctx context.Context, task *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			task.ModifyProperty(dependCountProperty{}, func(_ interface{}) interface{} {
				return uint(0)
			})

			getLock(task).Lock()
			defer getLock(task).Unlock()

			return next(ctx)
		},

		Depend: func(ctx context.Context, task, _ *taskset.Task, next func(ctx context.Context) taskset.Result) taskset.Result {
			var dependCount uint

			task.ModifyProperty(dependCountProperty{}, func(value interface{}) interface{} {
				dependCount = value.(uint)
				return dependCount + 1
			})
			if dependCount == 0 {
				getLock(task).Unlock()
			}

			defer func() {
				task.ModifyProperty(dependCountProperty{}, func(value interface{}) interface{} {
					dependCount = value.(uint) - 1
					return dependCount
				})
				if dependCount == 0 {
					getLock(task).Lock()
				}
			}()

			return next(ctx)
		},
	}
}
