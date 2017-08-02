package threadpool

import (
	"context"
	"testing"
	"time"
)

// TestThreadLifeCycle tests creating, starting, running a function and joining a thread
func TestThreadLifeCycle(t *testing.T) {
	var (
		tasks   int64
		testVal = make(chan int)
	)
	tasksAddr := &tasks
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	thread := NewOSThread(ctx, cancel, tasksAddr)
	thread.Start()
	thread.Exec(func() {
		testVal <- 1
	})
	if val := <-testVal; val != 1 {
		t.Errorf("expected tasks to equal 1, but found %v", val)
	}
	thread.Join()
}
