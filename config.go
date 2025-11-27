package bgtaskqueue

import "context"

// config holds the configuration for the TaskQueue.
type config struct {
	ctx                 context.Context
	consumeTimeSpeed    uint32                               // 消费速率，单位秒「间隔多久执行一轮」
	consumePoolCapacity uint32                               // 消费者携程池数量
	releaseTimeout      uint32                               // 释放协程池的超时时间，单位秒
	loadDataTimeSepeed  uint32                               // 加载数据的时间间隔，单位秒
	producerLen         uint32                               // 任务生产者通道长度
	mode                ConsumerDistributionMode             // 消费者分发模式
	consumeFuncs        []func(*Task) error                  // 消费者处理函数列表
	dataFuncs           []func(args ...any) ([]*Task, error) // 加载任务数据的函数
}

// Option is a functional option for configuring the task queue.
type Option func(*config)

// WithContext sets the context for the task queue.
func WithContext(ctx context.Context) Option {
	return func(c *config) {
		c.ctx = ctx

	}
}

// WithProducerLen sets the length of the producer channel.
func WithProducerLen(length uint32) Option {
	return func(c *config) {
		c.producerLen = length
	}
}

// WithConcurrentCount sets the concurrent execution count for the queue.
func WithConsumePoolCapacity(count uint32) Option {
	return func(c *config) {
		c.consumePoolCapacity = count
	}
}

// WithReleaseTimeout sets the release timeout for the goroutine pool.
func WithReleaseTimeout(timeout uint32) Option {
	return func(c *config) {
		c.releaseTimeout = timeout
	}
}

// WithConcurrentTimeSpeed sets the time interval for concurrent execution.
func WithConsumeTimeSpeed(speed uint32) Option {
	return func(c *config) {
		c.consumeTimeSpeed = speed
	}
}

// WithLoadDataTimeSpeed sets the time interval for loading data.
func WithLoadDataTimeSpeed(speed uint32) Option {
	return func(c *config) {
		c.loadDataTimeSepeed = speed
	}
}

// WithConsumerDistributionMode sets the consumer distribution mode.
func WithConsumerDistributionMode(mode ConsumerDistributionMode) Option {
	return func(c *config) {
		c.mode = mode
	}
}

// WithConsumeFuncs sets the consume functions for processing tasks.
func WithConsumeFuncs(funcs ...func(*Task) error) Option {
	return func(c *config) {
		c.consumeFuncs = funcs
	}
}

// WithDataFuncs sets the data functions for loading tasks.
func WithDataFuncs(funcs ...func(args ...any) ([]*Task, error)) Option {
	return func(c *config) {
		c.dataFuncs = funcs
	}
}

// defaultConfig provides the default configuration for the task queue.
func defaultConfig() *config {
	return &config{
		ctx:                 context.Background(),
		consumePoolCapacity: 5,
		releaseTimeout:      3,
		consumeTimeSpeed:    1,
		loadDataTimeSepeed:  5,
		producerLen:         0,
		mode:                ConsumerDistributionModeFanout,
	}
}

// applyOptions applies the given options to the default configuration.
func applyOptions(opts ...Option) *config {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
