package bgtaskqueue

type Task struct {
	UniqueTaskID string // 任务唯一标识
	Payload      any    // 任务负载数据
}
