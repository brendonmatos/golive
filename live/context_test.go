package live

import (
	"sync"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	c := NewContext()

	if c.Hooks == nil {
		t.Errorf("hooks are undefined")
	}
}

func TestContextEvents(t *testing.T) {
	c := NewContext()

	wg := sync.WaitGroup{}
	wg.Add(1)
	c.InjectHook("click", func() {
		wg.Done()
	})

	c.CallHook("click")

	if waitTimeout(&wg, time.Second) {
		t.Error("event was not emitted")
	}
}

func TestContextGlobalEvents(t *testing.T) {
	c := NewContext()

	wg := sync.WaitGroup{}
	wg.Add(1)
	c.InjectGlobalHook("click", func() {
		wg.Done()
	})

	c.CallHook("click")

	if waitTimeout(&wg, time.Second) {
		t.Error("event was not emitted")
	}
}

func TestContextGlobalEventsWithNested(t *testing.T) {
	c := NewContext()

	wg := sync.WaitGroup{}
	wg.Add(1)
	c.InjectGlobalHook("click", func() {
		wg.Done()
	})

	c2 := c.Child()

	c2.CallHook("click")

	if waitTimeout(&wg, time.Second) {
		t.Error("event was not emitted")
	}
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
