package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

var globalCloser = New() //nolint:gochecknoglobals // just for convenience

// Add adds closing function to Closer.
func Add(f ...func() error) {
	globalCloser.Add(f...)
}

// Wait waits for Closer to be done.
func Wait() {
	globalCloser.Wait()
}

// CloseAll calls all closing functions on Closer.
func CloseAll() {
	globalCloser.CloseAll()
}

type Closer struct {
	mu        sync.Mutex
	once      sync.Once
	done      chan struct{}
	functions []func() error
}

func New(signals ...os.Signal) *Closer {
	closer := Closer{done: make(chan struct{})}
	if len(signals) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, signals...)
			<-ch
			signal.Stop(ch)
			closer.CloseAll()
		}()
	}
	return &closer
}

func (c *Closer) Add(functions ...func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.functions = append(c.functions, functions...)
}

func (c *Closer) Wait() {
	<-c.done
}

// CloseAll calls all close function from Closer functions.
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)
		c.mu.Lock()
		functions := c.functions
		c.functions = nil
		c.mu.Unlock()

		// Close all resources concurrently
		errs := make(chan error, len(functions))
		for _, f := range functions {
			go func(f func() error) {
				errs <- f()
			}(f)
		}

		for err := range errs {
			log.Printf("err returned from closer functions: %v", err)
		}
	})
}
