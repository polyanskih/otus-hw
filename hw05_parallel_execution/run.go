package hw05parallelexecution

import (
	"errors"
	"sync"
)

var (
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	ErrInvalidWorkersCount = errors.New("invalid workers count")
)

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrInvalidWorkersCount
	}

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		next        int
		errorsCount int
	)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				if next == len(tasks) || (m > 0 && errorsCount >= m) {
					mu.Unlock()
					return
				}
				task := tasks[next]
				next++
				mu.Unlock()

				if err := task(); err != nil {
					mu.Lock()
					errorsCount++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	if m > 0 && errorsCount >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}
