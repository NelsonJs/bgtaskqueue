package bgtaskqueue

import (
	"container/list"
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	l := list.New()
	l.PushBack(&Task{UniqueTaskID: "1", Payload: "payload1"})
	l.PushBack(&Task{UniqueTaskID: "2", Payload: "payload2"})
	l.PushBack(&Task{UniqueTaskID: "3", Payload: "payload3"})

	e := l.Front()
	l.Remove(e)
	println(l.Len())
}

func TestA(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := New(
		WithContext(ctx),
		WithDataFuncs(dataSource, dataSource1),
		WithConsumeFuncs(do, do1),
		WithLoadDataTimeSpeed(1),
	)
	if err != nil {
		panic(err)
	}
	time.Sleep(12 * time.Second)
}

func dataSource(args ...any) (ary []*Task, err error) {
	for i := 0; i < 20; i++ {
		ary = append(ary, &Task{UniqueTaskID: fmt.Sprintf("%d", time.Now().UnixNano()+int64(i)), Payload: "payload0"})
	}
	return
}

func dataSource1(args ...any) (ary []*Task, err error) {
	ary = append(ary, &Task{UniqueTaskID: fmt.Sprintf("%d", time.Now().UnixNano()), Payload: "payload1"})
	return
}

func do(task *Task) error {
	println("do >>>>", task.UniqueTaskID, task.Payload.(string))
	return nil
}
func do1(task *Task) error {
	println("do1 >>>>", task.UniqueTaskID, task.Payload.(string))
	return nil
}
