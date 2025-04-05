package librarySync

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

func GetWait() (func(), func(), func()) {
	var wg sync.WaitGroup
	return func() {
		wg.Add(1)
	}, wg.Done, wg.Wait
}

func GetMutex() (func(), func()) {
	var d sync.Mutex
	return d.Lock, d.Unlock
}

func GetSemaphore(amount int) (func(), func()) {
	s := semaphore.NewWeighted(int64(amount))
	return func() { s.Acquire(context.Background(), 1) }, func() { s.Release(1) }
}

func GetOnce() func(func()) {
	var once sync.Once
	return func(f func()) {
		once.Do(f)
	}
}