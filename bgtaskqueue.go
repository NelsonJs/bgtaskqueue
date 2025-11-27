package bgtaskqueue

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
)

var (
	// ErrNoConsumeFunc indicates that no consume function was provided.
	ErrNoConsumeFunc = errors.New("no consume function provided")
	ErrNoDataFunc    = errors.New("no data function provided")
)

type TaskQueue struct {
	waitQueue   *list.List   // 待执行任务队列
	queueCount  atomic.Int32 // 等待队列任务数量
	runingCount atomic.Int32 // 正在执行的异步任务数量
	gPool       *ants.Pool   // 携程池
	cfg         *config
	taskPool    sync.Map  // 任务池
	producer    chan Task // 额外的任务生产者通道
	queueMu     sync.Mutex
	consumeMu   sync.RWMutex // 保护 cfg.consumeFuncs 的读写
	dataFuncsMu sync.RWMutex // 保护 cfg.dataFuncs 的读写
}

func New(opts ...Option) (*TaskQueue, error) {
	cfg := applyOptions(opts...)
	if len(cfg.consumeFuncs) == 0 {
		return nil, ErrNoConsumeFunc
	}
	if len(cfg.dataFuncs) == 0 {
		return nil, ErrNoDataFunc
	}
	ins := &TaskQueue{
		waitQueue: list.New(),
		producer:  make(chan Task, cfg.producerLen),
	}
	ins.cfg = cfg

	pool, err := ants.NewPool(int(cfg.consumePoolCapacity))
	if err != nil {
		return nil, err
	}
	ins.gPool = pool
	go ins.watchProducer()
	go ins.Load()
	go ins.consumer()
	return ins, nil
}

// watchProducer listens for tasks on the producer channel and adds them to the task queue.
func (tq *TaskQueue) watchProducer() {
	for {
		select {
		case <-tq.cfg.ctx.Done():
			return
		case task := <-tq.producer:
			tq.PutTask(&task)
		}
	}
}

// Load loads tasks into the queue using the provided task function.
func (tq *TaskQueue) Load() error {
	ticker := time.NewTicker(time.Duration(tq.cfg.loadDataTimeSepeed) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 检查任务等待队列是否有数据，没有则读取dataFunc函数获取数据加入到本地内存队列中
			if tq.queueCount.Load() > 0 {
				continue
			}
			tq.dataFuncsMu.RLock()
			dataFuncs := make([]func(args ...any) ([]*Task, error), len(tq.cfg.dataFuncs))
			copy(dataFuncs, tq.cfg.dataFuncs)
			tq.dataFuncsMu.RUnlock()
			// 读取数据并加入到任务队列中
			for _, dataFunc := range dataFuncs {
				tasks, err := dataFunc()
				if err != nil {
					return err
				}
				for _, v := range tasks {
					tq.PutTask(v)
				}
			}
		case <-tq.cfg.ctx.Done():
			tq.release()
			return nil
		}
	}
}

// consumer processes tasks from the wait queue based on the configured distribution mode.
func (tq *TaskQueue) consumer() error {
	ticker := time.NewTicker(time.Duration(tq.cfg.consumeTimeSpeed) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 检查是否有可执行任务
			canRunCount := int(tq.cfg.consumePoolCapacity) - int(tq.runingCount.Load())
			if canRunCount <= 0 || tq.queueCount.Load() == 0 {
				continue
			}

			// 读取当前消费函数快照（读锁，确保添加时不会并发写）
			tq.consumeMu.RLock()
			funcs := make([]func(*Task) error, len(tq.cfg.consumeFuncs))
			copy(funcs, tq.cfg.consumeFuncs)
			tq.consumeMu.RUnlock()

			for i := 0; i < canRunCount; i++ {
				tq.queueMu.Lock()
				if tq.queueCount.Load() == 0 {
					tq.queueMu.Unlock()
					break
				}
				element := tq.waitQueue.Front()
				task := element.Value.(*Task)
				tq.waitQueue.Remove(element)
				tq.queueMu.Unlock()
				tq.queueCount.Add(-1)

				// 检查任务是否在任务池中
				if _, ok := tq.taskPool.Load(task.UniqueTaskID); !ok {
					continue
				}
				tq.taskPool.Delete(task.UniqueTaskID)

				switch tq.cfg.mode {
				case ConsumerDistributionModeRoundRobin:
					for _, f := range funcs {
						tq.runingCount.Add(1)
						err := tq.gPool.Submit(func() {
							defer tq.runingCount.Add(-1)
							_ = f(task)
						})
						if err != nil {
							tq.runingCount.Add(-1)
							return err
						}
					}
				case ConsumerDistributionModeFanout:
					tq.runingCount.Add(1)
					err := tq.gPool.Submit(func() {
						defer tq.runingCount.Add(-1)
						for _, f := range funcs {
							_ = f(task)
						}
					})
					if err != nil {
						tq.runingCount.Add(-1)
						return err
					}
				}
			}

		case <-tq.cfg.ctx.Done():
			return nil
		}
	}
}

// AddConsumerFunc adds a new consume function to the task queue.
func (tq *TaskQueue) AddConsumerFunc(consumeFunc func(*Task) error) {
	tq.consumeMu.Lock()
	defer tq.consumeMu.Unlock()
	tq.cfg.consumeFuncs = append(tq.cfg.consumeFuncs, consumeFunc)
}

// AddDataFunc adds a new data function to the task queue.
func (tq *TaskQueue) AddDataFunc(dataFunc func(args ...any) ([]*Task, error)) {
	tq.dataFuncsMu.Lock()
	defer tq.dataFuncsMu.Unlock()
	tq.cfg.dataFuncs = append(tq.cfg.dataFuncs, dataFunc)
}

// ClearAbortPool clears all entries in the abort pool.
func (tq *TaskQueue) AbortTask(taskUniqueID string) {
	tq.taskPool.Delete(taskUniqueID)
}

// PutTask adds a new task item to the wait queue.
func (tq *TaskQueue) PutTask(item *Task) {
	tq.queueMu.Lock()
	defer tq.queueMu.Unlock()
	if item.UniqueTaskID == "" {
		return
	}
	tq.waitQueue.PushBack(item)
	tq.queueCount.Add(1)
	tq.addToTaskPool(item.UniqueTaskID)
}

// GetQueueLength returns the current length of the wait queue.
func (tq *TaskQueue) GetQueueLength() int {
	return tq.waitQueue.Len()
}

// AbortTask marks a task with the given uniqueTaskID to be aborted.
func (tq *TaskQueue) addToTaskPool(uniqueTaskID string) {
	tq.taskPool.Store(uniqueTaskID, struct{}{})
}

// GetProducerChan returns the producer channel.
func (tq *TaskQueue) GetProducerChan() chan Task {
	return tq.producer
}

func (tq *TaskQueue) release() {
	println("release")
	close(tq.producer)
	if tq.cfg.releaseTimeout <= 0 {
		tq.gPool.Release()
		return
	}
	tq.gPool.ReleaseTimeout(time.Duration(tq.cfg.releaseTimeout) * time.Second)
}
