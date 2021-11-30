package middlewares_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bennydictor/taskset"
	"github.com/bennydictor/taskset/middlewares"
)

func ExampleNewConcurrencyLimiter_sequential() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet(
		middlewares.NewConcurrencyLimiter(&sync.Mutex{}),
	)

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 1, nil
	})

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 2, nil
	})

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 3, nil
	})

	start := time.Now()
	taskSet.Start(ctx)
	taskSet.Wait(ctx)
	totalTime := time.Since(start)

	fmt.Printf("total time: %.0fs\n", totalTime.Seconds())
	// Output: total time: 6s
}

func ExampleNewConcurrencyLimiter_semaphore() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet(
		middlewares.NewConcurrencyLimiter(middlewares.NewSemaphore(2)),
	)

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 1, nil
	})

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 2, nil
	})

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 3, nil
	})

	start := time.Now()
	taskSet.Start(ctx)
	taskSet.Wait(ctx)
	totalTime := time.Since(start)

	fmt.Printf("total time: %.0fs\n", totalTime.Seconds())
	// Output: total time: 4s
}

func ExampleNewConcurrencyLimiter_mutualExclusion() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet(
		middlewares.NewConcurrencyLimiter(nil),
	)

	// A and B can't run concurrently
	var abMu sync.Mutex

	// C and D can't run concurrently
	var cdMu sync.Mutex

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 1, nil
	},
		middlewares.WithLock(&abMu),
	)

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 2, nil
	},
		middlewares.WithLock(&abMu),
	)

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 3, nil
	},
		middlewares.WithLock(&cdMu),
	)

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 4, nil
	},
		middlewares.WithLock(&cdMu),
	)

	start := time.Now()
	taskSet.Start(ctx)
	taskSet.Wait(ctx)
	totalTime := time.Since(start)

	fmt.Printf("total time: %.0fs\n", totalTime.Seconds())
	// Output: total time: 4s
}
