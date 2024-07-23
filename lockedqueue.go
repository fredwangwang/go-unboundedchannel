package unboundedchannel

import (
	"sync"
	"sync/atomic"
)

type lqNode[T any] struct {
	val   T
	nextp atomic.Pointer[lqNode[T]]
}

type LockedQueue[T any] struct {
	pool   sync.Pool
	closed atomic.Bool

	popCond sync.Cond
	head    *lqNode[T]

	pushMtx sync.Mutex
	tail    *lqNode[T]

	len atomic.Int64
}

func NewLockQueue[T any]() *LockedQueue[T] {
	var q LockedQueue[T]

	q.head = &lqNode[T]{}
	q.tail = q.head

	q.pool = sync.Pool{
		New: func() any {
			return &lqNode[T]{}
		},
	}
	q.popCond = *sync.NewCond(&sync.Mutex{})

	return &q
}

func (q *LockedQueue[T]) Len() int {
	return int(q.len.Load())
}

func (q *LockedQueue[T]) Pop() (T, bool, error) {
	var in [1]T
	n, err := q.popslice(in[:], false)
	return in[0], n == 1, err

	// return q.pop(false)
}

func (q *LockedQueue[T]) PopWait() (T, bool, error) {
	var in [1]T
	n, err := q.popslice(in[:], true)
	return in[0], n == 1, err

	// return q.pop(true)
}

func (q *LockedQueue[T]) PopSlice(in []T) (int, error) {
	return q.popslice(in, false)
}

func (q *LockedQueue[T]) PopSliceWait(in []T) (int, error) {
	return q.popslice(in, true)
}

// func (q *Queue2[T]) pop(wait bool) (T, bool, error) {
// 	q.cond.L.Lock()

// 	if !wait && q.len.Load() == 0 {
// 		q.cond.L.Unlock()
// 		var v T

// 		if q.closed.Load() {
// 			return v, false, ErrQueueClosed
// 		} else {
// 			return v, false, nil
// 		}
// 	}

// 	for q.len.Load() == 0 {
// 		q.cond.Wait()
// 		if q.closed.Load() {
// 			q.cond.L.Unlock()
// 			var v T

// 			return v, false, nil
// 		}
// 	}

// 	q.len.Add(-1)

// 	head := q.head
// 	next := head.next
// 	q.head = next

// 	v := next.val
// 	q.cond.L.Unlock()

// 	*head = qNode2[T]{}
// 	q.pool.Put(head)

// 	return v, true, nil
// }

func (q *LockedQueue[T]) popslice(in []T, wait bool) (int, error) {
	q.popCond.L.Lock()

	for q.len.Load() == 0 {
		if q.closed.Load() {
			q.popCond.L.Unlock()
			return 0, ErrQueueClosed
		}
		if !wait {
			q.popCond.L.Unlock()
			return 0, nil
		}

		q.popCond.Wait()
	}

	inlen := len(in)
	n := 0

	head := q.head
	next := head.nextp.Load()

	for next != nil && n < inlen {
		q.head = next
		q.len.Add(-1)

		v := next.val

		in[n] = v
		n += 1

		*head = lqNode[T]{}
		q.pool.Put(head)

		head = q.head
		next = head.nextp.Load()
	}

	q.popCond.L.Unlock()

	return n, nil
}

func (q *LockedQueue[T]) Push(val T) error {
	if q.closed.Load() {
		return ErrQueueClosed
	}

	node := q.pool.Get().(*lqNode[T])
	*node = lqNode[T]{val: val}

	var newLen int64

	q.pushMtx.Lock()

	q.tail.nextp.Store(node)
	q.tail = node
	newLen = q.len.Add(1)

	q.pushMtx.Unlock()

	// newLen == 1, original len == 0, which means the queue is empty
	// signal the consumer to take items.
	if newLen == 1 {
		// locking before signal the Cond.
		// Otherwise the signal could be lost if the consuming goroutine
		// is preempted while holding the lock but yet to wait for cond,
		// resulting eventual deadlock on the consumer (never signalled to wake up).
		q.popCond.L.Lock()
		q.popCond.Signal()
		q.popCond.L.Unlock()
	}

	return nil
}

func (q *LockedQueue[T]) Close() {
	q.closed.Store(true)
	q.popCond.Broadcast()
}
