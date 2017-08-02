// +build go1.7
package threadpool

import (
	"context"
	"testing"
	"time"
)

// TestThreadpoolLifecyle tests creating, starting, running a function and stopping a threadpool
func TestThreadpoolLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	tp, err := New(ctx, cancel, 2)
	if err != nil {
		t.Error(err)
	}
	tp.Start()
	resChan := make(chan struct{}, 2)
	testFunc := func() {
		resChan <- struct{}{}
	}
	for i := 2; i > 0; i-- {
		tp.Exec(testFunc)
	}
	var counter int
	for _ = range resChan {
		counter++
		if counter == 2 {
			return
		}
	}
	t.Error("did not receive two channel results")
}
