package gpool

import (
	"errors"
	"sync"
	"sync/atomic"
)

// Pool is a fixed-size goroutine pool that executes submitted tasks
// concurrently.
// Pool 是一个固定大小的协程池，用于并发执行提交的任务。
type Pool struct {
	size     uint64
	running  uint64
	taskPool chan func()

	mu      sync.Mutex
	isClose bool

	sync.WaitGroup
}

// NewTaskPool creates a pool with at most size worker goroutines. A size < 1 is
// treated as 1 to guarantee the pool can always make progress.
// NewTaskPool 创建一个最多包含 size 个 worker 的协程池。size < 1 时按 1 处理，保证池能正常推进。
func NewTaskPool(size int) *Pool {
	if size < 1 {
		size = 1
	}
	return &Pool{
		size:     uint64(size),
		taskPool: make(chan func()),
	}
}

// run starts a new worker goroutine.
// run 启动一个新的 worker 协程。
func (p *Pool) run() {
	p.incRunning()
	go func() {
		defer p.decRunning()
		for task := range p.taskPool {
			p.execute(task)
		}
	}()
}

// execute runs a task and always releases its WaitGroup slot, even when the task
// panics. Without the recover a single panicking task would kill the worker and
// leak the counter, causing Close to block forever.
// execute 执行任务并始终释放 WaitGroup 计数，即使任务发生 panic。
// 若没有 recover，一次 panic 会杀死 worker 并泄漏计数，导致 Close 永久阻塞。
func (p *Pool) execute(task func()) {
	defer p.Done()
	defer func() {
		_ = recover()
	}()
	task()
}

func (p *Pool) incRunning() { atomic.AddUint64(&p.running, 1) }
func (p *Pool) decRunning() { atomic.AddUint64(&p.running, ^uint64(0)) }

// Running returns the number of currently alive worker goroutines.
// Running 返回当前存活的 worker 协程数。
func (p *Pool) Running() uint64 {
	return atomic.LoadUint64(&p.running)
}

// GetSize returns the maximum number of workers.
// GetSize 返回 worker 数量上限。
func (p *Pool) GetSize() uint64 {
	return p.size
}

// Submit adds a task to the pool. It returns an error if the pool is closed or
// the task is nil.
// Submit 向池中提交任务。池已关闭或任务为 nil 时返回错误。
func (p *Pool) Submit(task func()) error {
	if task == nil {
		return errors.New("gpool: nil task")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClose {
		return errors.New("gpool: pool closed")
	}
	// Register before dispatching so the WaitGroup count matches the number of
	// in-flight tasks. Checking isClose first avoids leaking the counter.
	// 先注册再派发，使计数与在途任务数一致；先判断 isClose 可避免计数泄漏。
	p.Add(1)
	if p.Running() < p.GetSize() {
		p.run()
	}
	p.taskPool <- task
	return nil
}

// Close stops accepting new tasks, waits for all submitted tasks to finish and
// then terminates the workers. It is safe to call multiple times.
// Close 停止接收新任务，等待所有已提交任务完成后终止 worker，可重复调用。
func (p *Pool) Close() {
	p.mu.Lock()
	if p.isClose {
		p.mu.Unlock()
		return
	}
	p.isClose = true
	p.mu.Unlock()

	// Wait() blocks until every submitted task has run (Add happens in Submit,
	// Done after execution), so no busy-waiting is required.
	// Wait() 会阻塞到所有已提交任务执行完毕，因此无需忙等。
	p.Wait()
	close(p.taskPool)
}
