package rungroup

import (
	"context"
	"sync"
)

type Group struct {
	mu     sync.RWMutex
	actors []Actor
}

type Actor interface {
	Run(context.Context) error
}

type ActorFunc func(context.Context) error

func (fn ActorFunc) Run(ctx context.Context) error {
	return fn(ctx)
}

// Unit is a wrapper around
type Unit struct {
	actor  Actor
	wg     *sync.WaitGroup
	errout chan error
}

type runCtx struct {
	wg    *sync.WaitGroup
	units []*Unit
}

func (u *Unit) Run(ctx context.Context, errout chan error) {
	defer u.wg.Done()
	if err := u.actor.Run(ctx); err != nil {
		select {
		case <-ctx.Done():
		case errout <- err:
		}
	}
}

func wait(done chan struct{}, wg *sync.WaitGroup) {
	defer close(done)
	wg.Wait()
}

func (g *Group) Add(actors ...Actor) error {
	g.mu.Lock()
	g.actors = append(g.actors, actors...)
	g.mu.Unlock()
	return nil
}

func (g *Group) Run(ctx context.Context) <-chan error {
	// Copy over the necessary stuff from g, and create a separate
	// run ctx that handles the actual execution. This way
	// we doucple the public API from Group with the currently
	// running actors, thereby avoiding the need for synchronization
	g.mu.RLock()
	var rc runCtx
	rc.wg = &sync.WaitGroup{}
	for _, actor := range g.actors {
		rc.AddActor(actor)
	}
	g.mu.RUnlock()

	return rc.Run(ctx)
}

func (rc *runCtx) AddActor(a Actor) {
	rc.units = append(rc.units, &Unit{
		wg:    rc.wg,
		actor: a,
	})
}

func (rc *runCtx) Run(ctx context.Context) <-chan error {
	errs := make(chan error, len(rc.units))
	defer close(errs)

	for _, unit := range rc.units {
		unit.wg.Add(1)
		go unit.Run(ctx, errs)
	}

	done := make(chan struct{})
	go wait(done, rc.wg)

	select {
	case <-ctx.Done():
	case <-done:
	}

	return errs
}
