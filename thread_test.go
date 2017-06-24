package threadpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestThreadLifeCycle tests creating, starting, running a function and joining a thread
func TestThreadLifeCycle(t *testing.T) {
	var (
		tasks         int64
		atomicTestVal int64
	)
	tasksAddr := &tasks
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	thread := NewOSThread(ctx, cancel, tasksAddr)
	thread.Start()
	thread.Exec(func() {
		time.Sleep(time.Second)
		atomic.AddInt64(&atomicTestVal, 1)
		if val := atomic.LoadInt64(&atomicTestVal); val != 1 {
			t.Errorf("expected atomicTestVal to equal 1, but found %v", val)
		}
	})
	totalTasks := atomic.LoadInt64(tasksAddr)
	if totalTasks != 1 {
		t.Errorf("expected tasks to equal 1, but found %v", totalTasks)
	}
	thread.Join()
}
