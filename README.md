# BackgroupTaskQueue
Production and consumption library based on local memory.
Sometimes we need to load data from a database or other data sources periodically. To avoid frequent access to the data source, such as putting too much pressure on the database, we can read data from the data source in batches, process it, and then read it again. This library was created to handle this requirement.

## Compatibility
1. Supports multiple data providers, including those providing data using channels.
2. Supports multiple consumers and multiple consumption models.

## Usage
```
go get -u github.com/NelsonJs/bgtaskqueue
```
```
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err :=
		bgtaskqueue.New(
			bgtaskqueue.WithContext(ctx),
			bgtaskqueue.WithDataFuncs(dataSource, dataSource1),
			bgtaskqueue.WithConsumeFuncs(do, do1),
			bgtaskqueue.WithLoadDataTimeSpeed(1),
		)
	if err != nil {
		panic(err)
	}
	time.Sleep(12 * time.Second)
}

func dataSource(args ...any) (ary []*bgtaskqueue.Task, err error) {
	for i := 0; i < 20; i++ {
		ary = append(ary, &bgtaskqueue.Task{UniqueTaskID: fmt.Sprintf("%d", time.Now().UnixNano()+int64(i)), Payload: "payload0"})
	}
	return
}

func dataSource1(args ...any) (ary []*bgtaskqueue.Task, err error) {
	ary = append(ary, &bgtaskqueue.Task{UniqueTaskID: fmt.Sprintf("%d", time.Now().UnixNano()), Payload: "payload1"})
	return
}

func do(task *bgtaskqueue.Task) error {
	println("do >>>>", task.UniqueTaskID, task.Payload.(string))
	return nil
}
func do1(task *bgtaskqueue.Task) error {
	println("do1 >>>>", task.UniqueTaskID, task.Payload.(string))
	return nil
}

```
## ConsumerSupportedMode
｜mode|description|
|-----|-----------|
|ConsumerDistributionModeFanout|A set of data will be distributed to all consumers|
|ConsumerDistributionModeRoundRobin|One set of data will only be sent to one consumer|
## Examples
This case study includes some simple usage [examples](./example/).
## License
This software is licensed under the Apache 2 license
