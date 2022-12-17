package utils

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	size    uint64
	running uint64
	//工作池
	taskPool chan func()
	isClose  bool
	sync.Mutex
	sync.WaitGroup
}

func NewTaskPool(size int) *Pool {
	//初始化一个Pool
	return &Pool{
		size:     uint64(size),
		taskPool: make(chan func()),
	}
}
func (p *Pool) run() {
	p.incRunning() // 运行中的任务加一

	go func() {
		defer func() {
			p.decRunning() // task 结束, 运行中的任务减一
		}()
		for {
			select { // 阻塞等待任务
			case t, ok := <-p.taskPool: // 从 channel 中消费任务
				if !ok { // 如果 channel 被关闭, 结束 worker 运行
					return
				}
				// 执行任务
				t()
				p.Done()

			}
		}
	}()
}
func (p *Pool) incRunning() { // running + 1
	atomic.AddUint64(&p.running, 1)
}

func (p *Pool) decRunning() { // running - 1
	atomic.AddUint64(&p.running, ^uint64(0))
}
func (p *Pool) setStatus(status bool) bool {
	p.Lock()
	defer p.Unlock()
	if p.isClose == status {
		return false
	}
	p.isClose = status
	return true
}

// Running  获取Runing状态的协程数
func (p *Pool) Running() uint64 {
	return atomic.LoadUint64(&p.running)
}
func (p *Pool) GetSize() uint64 {
	return p.size
}

// Submit 提交任务到task通道中
func (p *Pool) Submit(task func()) error {
	p.Lock()
	defer p.Unlock()
	p.Add(1)
	if p.isClose {
		return errors.New("pool Colosed")
	}
	if p.Running() < p.GetSize() { // 如果task池满, 则不再创建 task
		// 启动一个 task
		p.run()
	}
	// 将task推入通道, 等待消费
	if !p.isClose {
		p.taskPool <- task
	}
	return nil
}

// Close 关闭通道释放资源,
func (p *Pool) Close() {
	// 设置 isColose 为true表示停止
	p.setStatus(true)
	// 阻塞等待所有任务被 worker 消费
	for len(p.taskPool) > 0 {
		// 防止等待任务清空 cpu 负载突然变大, 这里小睡一下
		time.Sleep(1e6)
	}
	//在这阻塞，待到所有协程任务处理完毕关闭通道结束
	p.Wait()
	close(p.taskPool)
}
