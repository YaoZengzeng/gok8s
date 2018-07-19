package waitgroup

import (
	"fmt"
	"sync"
)

type SafeWaitGroup struct {
	wg sync.WaitGroup
	mu sync.Mutex

	wait bool
}

func (wg *SafeWaitGroup) Add(delta int) error {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.wait && delta > 0 {
		return fmt.Errorf("failed to Add for wait has started")
	}

	wg.wg.Add(delta)

	return nil
}

func (wg *SafeWaitGroup) Done() {
	wg.wg.Done()
}

func (wg *SafeWaitGroup) Wait() {
	wg.mu.Lock()
	wg.wait = true
	wg.mu.Unlock()

	wg.wg.Wait()
}
