package taskset_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bennydictor/taskset"
)

func ExampleTaskSet() {
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
}

// A lazy task that no other task depends on will never run.
func ExampleTaskSet_NewLazy() {
	ctx := context.Background()

	taskSet := taskset.NewTaskSet()

	var mu sync.Mutex
	var tasksStarted []string
	taskStarted := func(task string) {
		mu.Lock()
		defer mu.Unlock()
		tasksStarted = append(tasksStarted, task)
	}

	taskA := taskSet.NewLazy(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		taskStarted("A")
		return 1, nil
	})

	taskB := taskSet.NewLazy(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		taskStarted("B")
		return 2, nil
	})

	taskSet.NewLazy(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		taskStarted("C")
		return 3, nil
	})

	taskSet.New(func(ctx context.Context, depend taskset.Depend) (interface{}, error) {
		taskStarted("D")

		if t := depend.ErrGroup(ctx, taskA, taskB); t != nil {
			return nil, depend(ctx, t).Err
		}

		a := depend(ctx, taskA).Value.(int)
		b := depend(ctx, taskB).Value.(int)

		return a + b, nil
	})

	taskSet.Start(ctx)
	taskSet.Wait(ctx)

	sort.Strings(tasksStarted)
	fmt.Println(strings.Join(tasksStarted, ", "))

	// Output: A, B, D
}
