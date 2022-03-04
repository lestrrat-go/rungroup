package rungroup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lestrrat-go/rungroup"
	"github.com/stretchr/testify/assert"
)

func TestRunGroup(t *testing.T) {
	var g rungroup.Group

	for i := 0; i < 10; i++ {
		i := i
		g.Add(rungroup.ActorFunc(func(ctx context.Context) error {
			if i%2 == 1 {
				return fmt.Errorf(`%d`, i)
			}
			return nil
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := g.Run(ctx)

	time.Sleep(time.Second)
	cancel()

	if !assert.Len(t, err, 5) {
		return
	}
}

func TestWait(t *testing.T) {
	var cleanupCalled bool
	var g rungroup.Group
	g.Add(rungroup.ActorFunc(func(ctx context.Context) error {
		// This goroutine waits indefinitely until canceled
		<-ctx.Done()
		cleanupCalled = true
		return nil
	}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if !assert.NoError(t, <-g.Run(ctx), `g.Run() should succeed`) {
		return
	}

	if !assert.True(t, cleanupCalled, `cleanup should have been called`) {
		return
	}
}
