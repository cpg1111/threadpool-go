package threadpool

import (
	"context"
	"runtime"
	"sync/atomic"
)

type OSThread struct {
	exec       chan func()
	ctx        context.Context
	cancel     context.CancelFunc
	totalTasks *int64
	localTasks *int64
}

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

func (o *OSThread) Start() {
	go o.start()
}

func (o *OSThread) Exec(fn func()) {
	atomic.AddInt64(o.localTasks, 1)
	o.exec <- fn
}

func (o *OSThread) Join() {
	for _ = range o.ctx.Done() {
		tasks := atomic.LoadInt64(o.localTasks)
		if tasks == 0 {
			break
		}
	}
	o.cancel()
}

func (o *OSThread) Stop() {
	o.cancel()
}
