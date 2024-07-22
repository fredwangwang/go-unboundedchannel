package unboundedchannel

import (
	"errors"
	"sync"
	"sync/atomic"
)

// unused, found locked queue provide better perf actually..

var ErrQueueClosed = errors.New("queue already closed")

type qNode[T any] struct {
	val  T
	next atomic.Pointer[qNode[T]]
}

type Queue[T any] struct {
	dummy qNode[T]

	head atomic.Pointer[qNode[T]]
	tail atomic.Pointer[qNode[T]]
	len  atomic.Int64

	pool *sync.Pool

	cond *sync.Cond

	closed atomic.Bool
}

func NewQueue[T any]() *Queue[T] {
	var q Queue[T]

	q.head.Store(&q.dummy)
	q.tail.Store(q.head.Load())

	q.pool = &sync.Pool{
		New: func() any {
			return &qNode[T]{}
		},
	}

	q.cond = sync.NewCond(&sync.Mutex{})

	return &q
}

func (q *Queue[T]) Len() int {
	return int(q.len.Load())
}

func (q *Queue[T]) Pop() (T, bool, error) {
	return q.pop(false)
}

func (q *Queue[T]) PopWait() (T, bool, error) {
	return q.pop(true)
}

func (q *Queue[T]) pop(wait bool) (T, bool, error) {
	for {
		head := q.head.Load()
		next := head.next.Load()
		if next != nil {
			if q.head.CompareAndSwap(head, next) {
				q.len.Add(-1)

				v := next.val

				*head = qNode[T]{}
				q.pool.Put(head)

				return v, true, nil
			} else {
				continue
			}
		} else {
			if q.closed.Load() {
				var v T
				return v, false, ErrQueueClosed
			}

			if wait {
				q.cond.L.Lock()
				if q.len.Load() == 0 {
					q.cond.Wait()
				}
				q.cond.L.Unlock()
				continue
			}

			var v T
			return v, false, nil
		}
	}
}

func (q *Queue[T]) Push(val T) error {
	if q.closed.Load() {
		return ErrQueueClosed
	}

	node := q.pool.Get().(*qNode[T])
	*node = qNode[T]{val: val}

	for {
		tail := q.tail.Load()
		if tail.next.CompareAndSwap(nil, node) {
			q.tail.Store(node)

			q.len.Add(1)
			q.cond.Signal()

			return nil
		} else {
			continue
		}
	}
}

func (q *Queue[T]) Close() {
	q.closed.Store(true)
	q.cond.Broadcast()
}
