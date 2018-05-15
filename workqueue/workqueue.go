package workqueue

import (
	"sync"
)

type Interface interface {
	Add(item interface{})
	Len() int
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	ShutDown()
	ShuttingDown() bool
}

type empty struct{}
type t interface{}
type set map[t]empty

func (s set) insert(k interface{}) {
	s[k] = empty{}
}

func (s set) delete(k interface{}) {
	delete(s, k)
}

func (s set) has(k interface{}) bool {
	_, ok := s[k]
	return ok
}

type Type struct {
	queue []t

	dirty      set
	processing set

	shuttingDown bool

	cond *sync.Cond
}

func New() *Type {
	return &Type{
		dirty:      make(set),
		processing: make(set),
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}

func (t *Type) Add(item interface{}) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	if t.shuttingDown {
		return
	}

	if t.dirty.has(item) {
		return
	}

	t.dirty.insert(item)

	if !t.processing.has(item) {
		t.queue = append(t.queue, item)
	}

	t.cond.Signal()
}

func (t *Type) Get() (interface{}, bool) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	for len(t.queue) == 0 && t.shuttingDown != true {
		t.cond.Wait()
	}

	if len(t.queue) == 0 {
		return nil, true
	}

	var item interface{}

	item, t.queue = t.queue[0], t.queue[1:]
	t.dirty.delete(item)
	t.processing.insert(item)
	return item, false
}

func (t *Type) Done(item interface{}) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.processing.delete(item)
	if t.dirty.has(item) {
		t.queue = append(t.queue, item)
		t.cond.Signal()
	}
}

func (t *Type) ShutDown() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()
	t.shuttingDown = true
	t.cond.Broadcast()
}

func (t *Type) Len() int {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()
	return len(t.queue)
}

func (t *Type) ShuttingDown() bool {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()
	return t.shuttingDown
}
