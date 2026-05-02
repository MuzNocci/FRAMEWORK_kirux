package concurrency

import (
	"context"
	"sync"
)

type Group struct {
	wg   sync.WaitGroup
	errs []error
	mu   sync.Mutex
}

func (g *Group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(); err != nil {
			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}
	}()
}

func (g *Group) Wait() []error {
	g.wg.Wait()
	return g.errs
}

type WorkerPool struct {
	jobs    chan func()
	wg      sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, workers int) *WorkerPool {
	p := &WorkerPool{jobs: make(chan func(), workers*2)}
	for range workers {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-p.jobs:
					if !ok {
						return
					}
					job()
				}
			}
		}()
	}
	return p
}

func (p *WorkerPool) Submit(job func()) {
	p.jobs <- job
}

func (p *WorkerPool) Close() {
	close(p.jobs)
	p.wg.Wait()
}
