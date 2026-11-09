package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRunCases(t *testing.T) {
	defer goleak.VerifyNone(t)
	tests := []struct {
		name     string
		workers  int
		limit    int
		failures []bool
		count    int
		err      error
	}{
		{name: "empty tasks", workers: 3, limit: 1},
		{name: "one task", workers: 1, limit: 1, failures: []bool{false}, count: 1},
		{name: "more workers than tasks", workers: 8, limit: 1, failures: []bool{false, false}, count: 2},
		{name: "below error limit", workers: 1, limit: 3, failures: []bool{true, false, true}, count: 3},
		{
			name: "first error", workers: 1, limit: 1, failures: []bool{true, false, false},
			count: 1, err: ErrErrorsLimitExceeded,
		},
		{
			name: "errors separated by success", workers: 1, limit: 2, failures: []bool{true, false, true, false},
			count: 3, err: ErrErrorsLimitExceeded,
		},
		{
			name: "limit reached by last task", workers: 2, limit: 2, failures: []bool{true, true},
			count: 2, err: ErrErrorsLimitExceeded,
		},
		{name: "zero limit ignores errors", workers: 3, limit: 0, failures: []bool{true, true, false}, count: 3},
		{name: "negative limit ignores errors", workers: 3, limit: -1, failures: []bool{true, false, true}, count: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			counts := make([]int32, len(tc.failures))
			tasks := make([]Task, 0, len(tc.failures))
			for i, fail := range tc.failures {
				tasks = append(tasks, func() error {
					atomic.AddInt32(&counts[i], 1)
					if fail {
						return errors.New("task failed")
					}
					return nil
				})
			}

			err := Run(tasks, tc.workers, tc.limit)
			require.ErrorIs(t, err, tc.err)
			for i := range counts {
				var expected int32
				if i < tc.count {
					expected = 1
				}
				require.Equal(t, expected, atomic.LoadInt32(&counts[i]), "task %d", i)
			}
		})
	}
}

func TestRunInvalidWorkers(t *testing.T) {
	defer goleak.VerifyNone(t)
	for _, workers := range []int{0, -1} {
		t.Run(fmt.Sprint(workers), func(t *testing.T) {
			var calls int32
			tasks := []Task{func() error {
				atomic.AddInt32(&calls, 1)
				return nil
			}}
			require.EqualError(t, Run(tasks, workers, 1), "invalid workers count")
			require.Zero(t, atomic.LoadInt32(&calls))
		})
	}
}

func TestRunConcurrency(t *testing.T) {
	defer goleak.VerifyNone(t)
	tests := []struct {
		name    string
		workers int
		tasks   int
	}{
		{name: "one worker", workers: 1, tasks: 3},
		{name: "several workers", workers: 4, tasks: 12},
		{name: "more workers than tasks", workers: 8, tasks: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{}, tc.tasks)
			release := make(chan struct{})
			done := make(chan error, 1)
			counts := make([]int32, tc.tasks)
			var active int32
			tasks := make([]Task, 0, tc.tasks)
			for i := 0; i < tc.tasks; i++ {
				tasks = append(tasks, func() error {
					if count := atomic.AddInt32(&active, 1); count > int32(tc.workers) {
						t.Errorf("running %d tasks with %d workers", count, tc.workers)
					}
					started <- struct{}{}
					<-release
					atomic.AddInt32(&counts[i], 1)
					atomic.AddInt32(&active, -1)
					return nil
				})
			}
			go func() {
				done <- Run(tasks, tc.workers, 1)
			}()
			defer func() {
				close(release)
				require.NoError(t, waitForRun(t, done))
				require.Zero(t, atomic.LoadInt32(&active))
				for i := range counts {
					require.Equal(t, int32(1), atomic.LoadInt32(&counts[i]), "task %d", i)
				}
			}()

			for i := 0; i < min(tc.workers, tc.tasks); i++ {
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					t.Fatal("workers did not start tasks concurrently")
				}
			}
		})
	}
}

func TestRunWaitsAfterError(t *testing.T) {
	defer goleak.VerifyNone(t)
	started := make(chan struct{})
	failed := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	var finished int32
	tasks := []Task{
		func() error {
			close(started)
			<-release
			atomic.StoreInt32(&finished, 1)
			return nil
		},
		func() error {
			<-started
			close(failed)
			return errors.New("task failed")
		},
	}
	go func() {
		done <- Run(tasks, 2, 1)
	}()
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	select {
	case <-failed:
	case <-time.After(5 * time.Second):
		t.Fatal("error task did not run")
	}
	select {
	case err := <-done:
		t.Fatalf("Run returned before the running task finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	released = true
	require.ErrorIs(t, waitForRun(t, done), ErrErrorsLimitExceeded)
	require.Equal(t, int32(1), atomic.LoadInt32(&finished))
}

func waitForRun(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not finish")
		return nil
	}
}
