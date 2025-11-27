# BackgroupTaskQueue

基于本地内存的生产和消费库。

有时我们需要定期从数据库或其他数据源加载数据。为了避免频繁访问数据源（例如给数据库造成过大压力），我们可以分批从数据源读取数据，进行处理，处理完后再从数据源读取，以此循环往复。本库正是为了满足这一需求而创建的。

## 能力

1. 支持多个数据源提供者，包括使用channel提供数据。
2. 支持多个消费者，并支持选择消费模式

## 用法

```
go get -u github.com/NelsonJs/bgtaskqueue
```

```
_, err := bgtaskqueue.New(
			bgtaskqueue.WithContext(ctx),
			bgtaskqueue.WithDataFuncs(dataSource, dataSource1),
			bgtaskqueue.WithConsumeFuncs(do, do1),
			bgtaskqueue.WithLoadDataTimeSpeed(1),
		)
```

## 支持的消费模式

| 模式                               | 描述                                   |
| ---------------------------------- | -------------------------------------- |
| ConsumerDistributionModeFanout     | 一份数据将以广播的形式下发给全部消费者 |
| ConsumerDistributionModeRoundRobin | 一份数据只下发给一个消费者             |

## Examples

This case study includes some simple usage [examples](./example/).

## License

This software is licensed under the Apache 2 license
