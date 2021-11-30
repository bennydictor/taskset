package taskset_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bennydictor/taskset"
)

func ExampleDepend_ErrGroup() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet()

	taskA := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		return nil, errors.New("fail")
	})

	taskB := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 2, nil
	})

	taskC := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		if errTask := depend.ErrGroup(ctx, taskA, taskB); errTask != nil {
			return nil, depend(ctx, errTask).Err
		}

		a := depend(ctx, taskA).Value.(int)
		b := depend(ctx, taskB).Value.(int)

		return a + b, nil
	})

	start := time.Now()
	taskSet.Start(ctx)
	cResult := taskSet.Result(ctx, taskC)
	totalTime := time.Since(start)

	fmt.Printf("total time: %.0fs\n", totalTime.Seconds())
	if cResult.Err != nil {
		fmt.Println("C failed:", cResult.Err.Error())
	} else {
		fmt.Println("C result:", cResult.Value.(int))
	}

	// Output:
	// total time: 0s
	// C failed: fail
}

func ExampleDepend_SyncGroup() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet()

	taskA := taskSet.NewLazy(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return nil, errors.New("fail")
	})

	taskB := taskSet.NewLazy(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		time.Sleep(2 * time.Second)
		return 2, nil
	})

	taskC := taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		depend.SyncGroup(ctx, taskA, taskB)

		var a, b int
		if err := depend(ctx, taskA).Err; err == nil {
			a = depend(ctx, taskA).Value.(int)
		}
		if err := depend(ctx, taskB).Err; err == nil {
			b = depend(ctx, taskB).Value.(int)
		}

		return a + b, nil
	})

	start := time.Now()
	taskSet.Start(ctx)
	cResult := taskSet.Result(ctx, taskC)
	totalTime := time.Since(start)

	fmt.Printf("total time: %.0fs\n", totalTime.Seconds())
	if cResult.Err != nil {
		fmt.Println("C failed:", cResult.Err.Error())
	} else {
		fmt.Println("C result:", cResult.Value.(int))
	}

	// Output:
	// total time: 2s
	// C result: 2
}
