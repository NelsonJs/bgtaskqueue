package main

import (
	"context"
	"fmt"
	"time"

	"github.com/NelsonJs/bgtaskqueue"
)

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
