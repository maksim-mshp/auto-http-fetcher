package closer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.uber.org/multierr"
)

type closeFn func(ctx context.Context) error

type item struct {
	name string
	fn   closeFn
}

type Closer struct {
	log   *slog.Logger
	mu    sync.RWMutex
	items []item
}

func New(log *slog.Logger) *Closer {
	return &Closer{
		log: log,
		mu:  sync.RWMutex{},
	}
}

func (c *Closer) Add(name string, fn closeFn) error {
	if fn == nil {
		return fmt.Errorf("closer.Add: nil func")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = append(c.items, item{name: name, fn: fn})
	return nil
}

func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result error
	for _, item := range c.items {
		if err := item.fn(ctx); err != nil {
			c.log.Error("graceful shutdowm failed", "error", err)
			result = multierr.Append(result, err)
		}
	}
	c.log.Info("Shutdown stopped")
	return result
}
