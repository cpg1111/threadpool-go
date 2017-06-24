package threadpool

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

type Threadpool struct {
	exec       chan func()
	ctx        context.Context
	cancel     context.CancelFunc
	pool       *sync.Pool
	numThreads int
	totalTasks *int64
}

func New(ctx context.Context, cancel context.CancelFunc, num int) (*Threadpool, error) {
	if num >= runtime.GOMAXPROCS(0) {
		return nil, errors.New("threadpool size must be GOMAXPROCS - 1 at most")
	}
	if cancel == nil {
		ctx, cancel = context.WithCancel(ctx)
	}
	p := new(sync.Pool)
	tp := &Threadpool{
		exec:       make(chan func(), num),
		ctx:        ctx,
		cancel:     cancel,
		pool:       p,
		numThreads: num,
	}
	for i := 0; i < num; i++ {
		childCtx, childCancel := context.WithCancel(ctx)
		var localTaskAlloc int64
		t := &OSThread{
			exec:       make(chan func()),
			ctx:        childCtx,
			cancel:     childCancel,
			totalTasks: tp.totalTasks,
			localTasks: &localTaskAlloc,
		}
		t.Start()
		tp.pool.Put(t)
	}
	return tp, nil
}

func (t *Threadpool) start() {
	for {
		select {
		case _ = <-t.ctx.Done():
			for i := 0; i < t.numThreads; i++ {
				thread := t.pool.Get().(*OSThread)
				thread.Join()
			}
			return
		case fn := <-t.exec:
			thread := t.pool.Get().(*OSThread)
			defer t.pool.Put(thread)
			thread.Exec(fn)
			break
		}
	}
}

func (t *Threadpool) Start() {
	go t.start()
}

func (t *Threadpool) Stop() {
	t.cancel()
}

func (t *Threadpool) Exec(fn func()) {
	t.exec <- fn
}
