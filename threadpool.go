package threadpool

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

// Threadpool is a struct that manages a pool of OS threads
type Threadpool struct {
	exec       chan func()
	ctx        context.Context
	cancel     context.CancelFunc
	pool       *sync.Pool
	numThreads int
	totalTasks *int64
}

// New provides a constructor for a pointer to a new Threadpool.
// It will instantiate the pool and start its underlying threads.
// However, if a number of threads greater than or equal to GOMAXPROCS
// is requested, New will return an error instead.  The largest possible threadpool
// is GOMAXPROCS - 1, so that at a minimum one OS thread given to the Go runtime is still
// free to be scheduled by the Go runtime and execute the runtime code that would otherwise
// be blocked by code executing on the threads in the threadpool
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

// Start spawns the threadpool
func (t *Threadpool) Start() {
	go t.start()
}

// Stop will shutdown the threadpool.
// However, Threadpool.Stop() will call all child thread's Join()
// method, not thread.Stop()
func (t *Threadpool) Stop() {
	t.cancel()
}

// Exec queues a function for assignment to an OSThread.
// Once an OSThread is free to take the function, thread.Exec(fn)
// is called on the chosen thread.
func (t *Threadpool) Exec(fn func()) {
	t.exec <- fn
}
