// +build go 1.7
package threadpool

import (
	"context"
	"runtime"
	"sync/atomic"
)

// OSThread is a struct that provides metadata around an OS thread
type OSThread struct {
	exec       chan func()
	ctx        context.Context
	cancel     context.CancelFunc
	totalTasks *int64
	localTasks *int64
}

// NewOSThread is a constructor for a pointer to a new OSThread struct.
// It takes a context.Context, context.CancelFunc and a pointer to an int64.
// The context.Context and context.CancelFunc manage the lifecycle of the thread.
// The pointer to an int64 provides a counter for all tasks that *could* be passed to
// the OS thread.
func NewOSThread(ctx context.Context, cancel context.CancelFunc, tasks *int64) *OSThread {
	if cancel == nil {
		ctx, cancel = context.WithCancel(ctx)
	}
	var localTaskAlloc int64
	return &OSThread{
		exec:       make(chan func()),
		ctx:        ctx,
		cancel:     cancel,
		totalTasks: tasks,
		localTasks: &localTaskAlloc,
	}
}

func (o *OSThread) start() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for {
		select {
		case _ = <-o.ctx.Done():
			return
		case fn := <-o.exec:
			fn()
			atomic.AddInt64(o.localTasks, -1)
			break
		}
	}
}

// Start spawns the OS thread
func (o *OSThread) Start() {
	go o.start()
}

// Exec queues up a function for execution on the thread
func (o *OSThread) Exec(fn func()) {
	atomic.AddInt64(o.localTasks, 1)
	o.exec <- fn
}

// Join provides a graceful shutdown of the thread, in that it will wait
// for all queued functions for this thread to finish before unlocking the
// goroutine context off the thread
func (o *OSThread) Join() {
	for _ = range o.ctx.Done() {
		tasks := atomic.LoadInt64(o.localTasks)
		if tasks == 0 {
			break
		}
	}
	o.cancel()
}

// Stop is similiar to Join, however Stop is a hard shutdown that
// will immediately stop the thread, and unlock the goroutine
func (o *OSThread) Stop() {
	o.cancel()
}
