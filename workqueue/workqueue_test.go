package workqueue

import (
	"sync"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	q := New()

	const producer = 50
	producerWG := sync.WaitGroup{}
	producerWG.Add(producer)
	for i := 0; i < producer; i++ {
		go func(i int) {
			defer producerWG.Done()
			for j := 0; j < 50; j++ {
				q.Add(i)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	const consumer = 10
	consumerWG := sync.WaitGroup{}
	consumerWG.Add(consumer)
	for i := 0; i < consumer; i++ {
		go func(i int) {
			defer consumerWG.Done()
			for {
				item, quit := q.Get()
				if item == "added after shutdown!" {
					t.Errorf("Got an item added after shutdown")
				}
				if quit {
					break
				}
				t.Logf("Worker %v: begin processing %v", i, item)
				time.Sleep(3 * time.Millisecond)
				t.Logf("Worker %v: done processing %v", i, item)
				q.Done(item)
			}
		}(i)
	}

	producerWG.Wait()
	q.ShutDown()
	q.Add("added after shutdown!")
	consumerWG.Wait()
}

func TestLen(t *testing.T) {
	q := New()
	q.Add("foo")
	if e, a := 1, q.Len(); e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
	q.Add("bar")
	if e, a := 2, q.Len(); e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
	q.Add("foo")
	if e, a := 2, q.Len(); e != a {
		t.Errorf("Expected %v, got %v", e, a)
	}
}

func TestReinsert(t *testing.T) {
	q := New()

	q.Add("foo")

	i, _ := q.Get()
	if i != "foo" {
		t.Errorf("Expected %v, got %v", "foo", i)
	}

	q.Add(i)

	q.Done(i)

	i, _ = q.Get()
	if i != "foo" {
		t.Errorf("Expected %v, got %v", "foo", i)
	}

	q.Done(i)

	if a := q.Len(); a != 0 {
		t.Errorf("Expected queue to be empty. Has %v items", a)
	}
}
