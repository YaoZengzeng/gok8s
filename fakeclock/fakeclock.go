package clock

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
	After(time.Duration) <-chan time.Time
	NewTimer(time.Duration) Timer
	Sleep(time.Duration)
	NewTicker(time.Duration) Ticker
}

type FakeClock struct {
	lock sync.RWMutex
	time time.Time

	// waiters are waiting for the fake time to pass their specified time
	waiters []*fakeClockWaiter
}

type fakeClockWaiter struct {
	targetTime    time.Time
	stepInterval  time.Duration
	skipIfBlocked bool
	destChan      chan time.Time
	fired         bool
}

func NewFakeClock(t time.Time) *FakeClock {
	return &FakeClock{
		time: t,
	}
}

func (f *FakeClock) Now() time.Time {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.time
}

func (f *FakeClock) Since(ts time.Time) time.Duration {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.time.Sub(ts)
}

func (f *FakeClock) After(d time.Duration) <-chan time.Time {
	f.lock.RLock()
	defer f.lock.RUnlock()
	stopTime := f.time.Add(d)
	ch := make(chan time.Time, 1)
	f.waiters = append(f.waiters, &fakeClockWaiter{
		targetTime: stopTime,
		destChan:   ch,
	})
	return ch
}

func (f *FakeClock) Tick(d time.Duration) <-chan time.Time {
	f.lock.Lock()
	defer f.lock.Unlock()
	tickTime := f.time.Add(d)
	ch := make(chan time.Time, 1)
	f.waiters = append(f.waiters, &fakeClockWaiter{
		targetTime:    tickTime,
		stepInterval:  d,
		skipIfBlocked: true,
		destChan:      ch,
	})

	return ch
}

// Move clock by Duration, notify anyone that's called After, Tick, or NewTimer
func (f *FakeClock) Step(d time.Duration) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.setTimeLocked(f.time.Add(d))
}

func (f *FakeClock) SetTime(t time.Time) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.setTimeLocked(t)
}

// Actually changes the time and checks any waiters. f must be write-locked.
func (f *FakeClock) setTimeLocked(t time.Time) {
	f.time = t
	newWaiters := make([]*fakeClockWaiter, 0, len(f.waiters))
	for i := range f.waiters {
		w := f.waiters[i]
		if !w.targetTime.After(t) {
			if w.skipIfBlocked {
				select {
				case w.destChan <- t:
					w.fired = true
				default:
				}
			} else {
				w.destChan <- t
				w.fired = true
			}

			if w.stepInterval > 0 {
				for !w.targetTime.After(t) {
					w.targetTime = w.targetTime.Add(w.stepInterval)
				}
				newWaiters = append(newWaiters, w)
			}
		} else {
			newWaiters = append(newWaiters, w)
		}
	}
	f.waiters = newWaiters
}

func (f *FakeClock) HasWaiters() bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return len(f.waiters) > 0
}

func (f *FakeClock) Sleep(d time.Duration) {
	f.Step(d)
}

func (f *FakeClock) NewTimer(d time.Duration) Timer {
	f.lock.Lock()
	defer f.lock.Unlock()
	stopTime := f.time.Add(d)
	ch := make(chan time.Time, 1)
	timer := &fakeTimer{
		fakeClock: f,
		waiter: &fakeClockWaiter{
			targetTime: stopTime,
			destChan:   ch,
		},
	}
	f.waiters = append(f.waiters, timer.waiter)
	return timer
}

func (f *FakeClock) NewTicker(d time.Duration) Ticker {
	f.lock.Lock()
	defer f.lock.Unlock()
	tickTime := f.time.Add(d)
	ch := make(chan time.Time, 1)
	f.waiters = append(f.waiters, &fakeClockWaiter{
		targetTime:    tickTime,
		stepInterval:  d,
		skipIfBlocked: true,
		destChan:      ch,
	})

	return &fakeTicker{
		c: ch,
	}
}

type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

var (
	_ = Timer(&fakeTimer{})
)

type fakeTimer struct {
	fakeClock *FakeClock
	waiter    *fakeClockWaiter
}

// C returns the channel that notifies when this timer has fired.
func (f *fakeTimer) C() <-chan time.Time {
	return f.waiter.destChan
}

// Stop stops the timer and returns true if the timer has not yet fired, or false otherwise.
func (f *fakeTimer) Stop() bool {
	f.fakeClock.lock.Lock()
	defer f.fakeClock.lock.Unlock()

	newWaiters := make([]*fakeClockWaiter, 0, len(f.fakeClock.waiters))
	for i := range f.fakeClock.waiters {
		w := f.fakeClock.waiters[i]
		if w != f.waiter {
			newWaiters = append(newWaiters, w)
		}
	}

	f.fakeClock.waiters = newWaiters

	return !f.waiter.fired
}

// Reset resets the timer to the fake clock's "now" + d. It returns true if the timer has not yet
// fired, or false otherwise.
func (f *fakeTimer) Reset(d time.Duration) bool {
	f.fakeClock.lock.Lock()
	defer f.fakeClock.lock.Unlock()

	active := !f.waiter.fired

	f.waiter.fired = false
	f.waiter.targetTime = f.fakeClock.time.Add(d)

	return active
}

type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type fakeTicker struct {
	c <-chan time.Time
}

func (t *fakeTicker) C() <-chan time.Time {
	return t.c
}

func (t *fakeTicker) Stop() {
}
