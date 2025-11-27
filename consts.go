package bgtaskqueue

type ConsumerDistributionMode int

const (
	ConsumerDistributionModeFanout     ConsumerDistributionMode = iota + 1 // 广播模式
	ConsumerDistributionModeRoundRobin                                     // 轮询模式
)
