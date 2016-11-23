package mqtt

import (
	"container/list"
	"sync"
)

type Queue struct {
	queue *list.List
	lock sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		queue: list.New(),
		lock: sync.RWMutex{},
	}
}

func (q *Queue) EnqueueMessage(msg interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.queue.PushBack(msg)
}
func (q *Queue) DequeueMessage() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.queue.Len() > 0 {
		el := q.queue.Front()
		v := q.queue.Remove(el)
		return v
	} else {
		return nil
	}
}
func (q *Queue) Size() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.queue.Len()
}
