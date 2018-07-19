package waitgroup

import (
	"testing"
)

func TestBasicWaitGroup(t *testing.T) {
	w1 := &SafeWaitGroup{}
	w2 := &SafeWaitGroup{}

	n := 20
	w1.Add(n)
	w2.Add(n)
	exitCh := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		go func() {
			w1.Done()
			w2.Wait()
			exitCh <- struct{}{}
		}()
	}

	w1.Wait()
	for i := 0; i < n; i++ {
		select {
		case <- exitCh:
			t.Fatalf("SafeWaitGroup released group too soon")
		default:
		}
		w2.Done()
	}
	for i := 0; i < n; i++ {
		<- exitCh
	}
}

func TestWaitGroupFail(t *testing.T) {
	w := &SafeWaitGroup{}
	w.Add(1)
	w.Done()
	w.Wait()
	if err := w.Add(1); err == nil {
		t.Fatalf("should return error when add positive after Wait")
	}
}
