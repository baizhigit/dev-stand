package closer

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
)

type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			c.CloseAll()
		}()
	}
	return c
}

func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

func (c *Closer) Wait() {
	<-c.done
}

func (c *Closer) CloseAll() {
	c.once.Do(func() { // sync.Once: safe to call from multiple goroutines simultaneously
		defer close(c.done)

		// Copy under lock then release — prevents deadlock if a closer func
		// calls Add() (which also needs the lock)
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		errs := make(chan error, len(funcs)) // buffered: goroutines never block on send
		for _, f := range funcs {
			go func(f func() error) {
				defer func() {
					// Panic recovery BEFORE wg.Done equivalent —
					// send result to channel, THEN signal completion
					if r := recover(); r != nil {
						errs <- fmt.Errorf("panic in closer: %v", r)
						return
					}
				}()
				errs <- f()
			}(f)
		}

		for i := 0; i < len(funcs); i++ {
			if err := <-errs; err != nil {
				slog.Error("closer func error", "err", err)
			}
		}
	})
}
