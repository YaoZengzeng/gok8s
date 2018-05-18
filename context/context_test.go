package context

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestBackground(t *testing.T) {
	c := Background()
	if c == nil {
		t.Fatalf("Background returned nil")
	}
	select {
	case x := <-c.Done():
		t.Errorf("<-c.Done() == %v want nothing (it should block)", x)
	default:
	}
}

func TestTODO(t *testing.T) {
	c := TODO()
	if c == nil {
		t.Fatalf("TODO returned nil")
	}
	select {
	case x := <-c.Done():
		t.Errorf("<-c.Done() == %v want nothing (it should block)", x)
	default:
	}
}

// otherContext is a Context that's not one of the types defined in context.go
// This let us test code paths that differ based on the underlying type of the
// Context.
type otherContext struct {
	Context
}

func TestWithCancel(t *testing.T) {
	c1, cancel := WithCancel(Background())

	o := otherContext{c1}
	c2, _ := WithCancel(o)
	contexts := []Context{c1, o, c2}

	for i, c := range contexts {
		if d := c.Done(); d == nil {
			t.Errorf("c[%d].Done() == %v want non-nil", i, d)
		}
		if e := c.Err(); e != nil {
			t.Errorf("c[%d].Err() == %v want nil", i, e)
		}

		select {
		case x := <-c.Done():
			t.Errorf("<-c.Done() == %v want nothing (it should block)", x)
		default:
		}
	}

	cancel()
	time.Sleep(100 * time.Millisecond) // let cancelation propagate

	for i, c := range contexts {
		select {
		case <-c.Done():
		default:
			t.Errorf("<-c[%d].Done() blocked, but shouldn't have", i)
		}
		if e := c.Err(); e != Canceled {
			t.Errorf("c[%d].Err() == %v want %v", i, e, Canceled)
		}
	}
}

func TestParentFinishesChild(t *testing.T) {
	// Context tree:
	// parent -> cancelChild
	// parent -> valueChild -> timerChild
	parent, cancel := WithCancel(Background())
	cancelChild, stop := WithCancel(parent)
	defer stop()
	timerChild, stop := WithTimeout(parent, 10000*time.Hour)
	defer stop()

	select {
	case x := <-parent.Done():
		t.Errorf("<-parent.Done() == %v want nothing (it should block)", x)
	case x := <-cancelChild.Done():
		t.Errorf("<-cancelChild.Done() == %v want nothing (it should block)", x)
	case x := <-timerChild.Done():
		t.Errorf("<-timerChild.Done() == %v want nothing (it should block)", x)
	default:
	}

	// The parent's children should contain the two cancelable children.
	pc := parent.(*cancelCtx)
	cc := cancelChild.(*cancelCtx)
	tc := timerChild.(*timerCtx)
	pc.mu.Lock()
	if len(pc.children) != 2 || !contains(pc.children, cc) || !contains(pc.children, tc) {
		t.Errorf("bad linkage: pc.children = %v, want %v and %v",
			pc.children, cc, tc)
	}
	pc.mu.Unlock()

	if p, ok := parentCancelCtx(cc.Context); !ok || p != pc {
		t.Errorf("bad linkage: parentCancelCtx(cancelChild.Context) = %v, %v want %v, true", p, ok, pc)
	}
	if p, ok := parentCancelCtx(tc.Context); !ok || p != pc {
		t.Errorf("bad linkage: parentCancelCtx(timerChild.Context) = %v, %v want %v, true", p, ok, pc)
	}

	cancel()

	pc.mu.Lock()
	if len(pc.children) != 0 {
		t.Errorf("pc.cancel didn't clear pc.children = %v", pc.children)
	}
	pc.mu.Unlock()

	// parent and children should all be finished.
	check := func(ctx Context, name string) {
		select {
		case <-ctx.Done():
		default:
			t.Errorf("<-%s.Done() blocked, but shouldn't have", name)
		}
		if e := ctx.Err(); e != Canceled {
			t.Errorf("%s.Err() == %v want %v", name, e, Canceled)
		}
	}
	check(parent, "parent")
	check(cancelChild, "cancelChild")
	check(timerChild, "timerChild")

	// WithCancel should return a canceled context on a canceled parent.
	precanceledChild, _ := WithTimeout(parent, 10000*time.Hour)
	select {
	case <-precanceledChild.Done():
	default:
		t.Errorf("<-precanceledChild.Done() blocked, but shouldn't have")
	}
	if e := precanceledChild.Err(); e != Canceled {
		t.Errorf("precanceledChild.Err() == %v want %v", e, Canceled)
	}
}

func contains(m map[canceler]struct{}, key canceler) bool {
	_, ret := m[key]
	return ret
}

func TestSimultaneousCancels(t *testing.T) {
	root, cancel := WithCancel(Background())
	m := map[Context]CancelFunc{root: cancel}
	q := []Context{root}
	// Create a tree of contexts.
	for len(q) != 0 && len(m) < 100 {
		parent := q[0]
		q = q[1:]
		for i := 0; i < 4; i++ {
			ctx, cancel := WithCancel(parent)
			m[ctx] = cancel
			q = append(q, ctx)
		}
	}
	// Start all the cancels in a random order.
	var wg sync.WaitGroup
	wg.Add(len(m))
	for _, cancel := range m {
		go func(cancel CancelFunc) {
			cancel()
			wg.Done()
		}(cancel)
	}
	// Wait on all the contexts in a random order.
	for ctx := range m {
		select {
		case <-ctx.Done():
		case <-time.After(1 * time.Second):
			buf := make([]byte, 10<<10)
			n := runtime.Stack(buf, true)
			t.Fatalf("timed out waiting for <-ctx.Done(); stacks:\n%s", buf[:n])
		}
	}
	// Wait for all the cancel functions to return.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		buf := make([]byte, 10<<10)
		n := runtime.Stack(buf, true)
		t.Fatalf("timed out waiting for cancel functions; stacks:\n%s", buf[:n])
	}
}

func TestInterlockedCancels(t *testing.T) {
	parent, cancelParent := WithCancel(Background())
	child, cancelChild := WithCancel(parent)
	go func() {
		parent.Done()
		cancelChild()
	}()
	cancelParent()
	select {
	case <-child.Done():
	case <-time.After(1 * time.Second):
		buf := make([]byte, 10<<10)
		n := runtime.Stack(buf, true)
		t.Fatalf("timed out waiting for child.Done(); stacks:\n%s", buf[:n])
	}
}

func TestDeadline(t *testing.T) {
	c, _ := WithDeadline(Background(), time.Now().Add(50*time.Millisecond))
	testDeadline(c, "WithDeadline", time.Second, t)

	c, _ = WithDeadline(Background(), time.Now().Add(50*time.Millisecond))
	o := otherContext{c}
	testDeadline(o, "WithDeadline+otherContext", time.Second, t)

	c, _ = WithDeadline(Background(), time.Now().Add(50*time.Millisecond))
	o = otherContext{c}
	c, _ = WithDeadline(o, time.Now().Add(4*time.Second))
	testDeadline(c, "WithDeadline+otherContext+WithDeadline", 2*time.Second, t)

	c, _ = WithDeadline(Background(), time.Now().Add(-time.Millisecond))
	testDeadline(c, "WithDeadline+inthepast", time.Second, t)

	c, _ = WithDeadline(Background(), time.Now())
	testDeadline(c, "WithDeadline+now", time.Second, t)
}

func testDeadline(c Context, name string, failAfter time.Duration, t *testing.T) {
	select {
	case <-time.After(failAfter):
		t.Fatalf("%s: context should have timed out", name)
	case <-c.Done():
	}
	if e := c.Err(); e != DeadlineExceeded {
		t.Errorf("%s: c.Err() == %v; want %v", name, e, DeadlineExceeded)
	}
}

func TestTimeout(t *testing.T) {
	c, _ := WithTimeout(Background(), 50*time.Millisecond)
	testDeadline(c, "WithTimeout", time.Second, t)

	c, _ = WithTimeout(Background(), 50*time.Millisecond)
	o := otherContext{c}
	testDeadline(o, "WithTimeout+otherContext", time.Second, t)

	c, _ = WithTimeout(Background(), 50*time.Millisecond)
	o = otherContext{c}
	c, _ = WithTimeout(o, 3*time.Second)
	testDeadline(c, "WithTimeout+otherContext+WithTimeout", 2*time.Second, t)
}
