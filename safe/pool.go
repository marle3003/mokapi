package safe

import (
	"context"
	"sync"
)

type Pool struct {
	waitGroup sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewPool(parentCtx context.Context) *Pool {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Pool{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (p *Pool) Go(f func(ctx context.Context)) {
	p.waitGroup.Add(1)
	go func() {
		defer func() {
			p.waitGroup.Done()
		}()
		f(p.ctx)
	}()
}

func (p *Pool) Stop() {
	p.cancel()
	p.waitGroup.Wait()
}
