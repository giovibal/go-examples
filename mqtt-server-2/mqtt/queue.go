package mqtt

import (
	"container/list"
	"log"
)

type Queue struct {
	queue             *list.List
}

func NewQueue() *Queue {
	return &Queue{
		queue: list.New(),
	}
}

func (q *Queue) EnqueueMessage(msg interface{}) {
	q.queue.PushBack(msg)
	log.Printf("Enqueued (Queue size: %s)\n", q.queue.Len())
}
func (q *Queue) DequeueMessage() interface{} {
	if q.queue.Len() > 0 {
		el := q.queue.Front()
		v := q.queue.Remove(el)
		log.Printf("Dequeued (Queue size: %v)\n", q.queue.Len())
		return v
	} else {
		log.Printf("Dequeued (Queue size: %v)\n", q.queue.Len())
		return nil
	}
}
