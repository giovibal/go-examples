package mqtt

import (
	"container/list"
)

type Queue struct {
	queue *list.List
}

func NewQueue() *Queue {
	return &Queue{
		queue: list.New(),
	}
}

func (q *Queue) EnqueueMessage(msg interface{}) {
	//q.queue.PushBack(msg)
}
func (q *Queue) DequeueMessage() interface{} {
	if q.queue.Len() > 0 {
		el := q.queue.Front()
		v := q.queue.Remove(el)
		return v
	} else {
		return nil
	}
}
func (q *Queue) Size() int {
	return q.queue.Len()
}
