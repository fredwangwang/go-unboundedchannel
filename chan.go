package unboundedchannel

import (
	"log"
	"sync/atomic"
)

type UnboundedChan[T any] struct {
	q  *LockedQueue[T]
	ch chan T

	closed atomic.Bool
}

func NewUnboundedChan[T any](initSize int) *UnboundedChan[T] {
	q := NewLockQueue[T]()
	ch := make(chan T, initSize)

	res := &UnboundedChan[T]{
		q:  q,
		ch: ch,
	}

	go func() {
		buf := make([]T, initSize)
		for {
			n, err := q.PopSliceWait(buf)
			if err == ErrQueueClosed {
				close(ch)
				return
			}

			for i := 0; i < n; i++ {
				ch <- buf[i]
			}
		}
	}()

	return res
}

func (uc *UnboundedChan[T]) Push(t T) {
	if !uc.closed.Load() {
		uc.q.Push(t)
	} else {
		log.Println("FATAL send on closed channel")
		panic("send on closed channel")
	}
}

// Len returns the total items in the channel.
// Because of the internal impl using a buffer, the exact number can be higher, and up to Len + initSize
func (uc *UnboundedChan[T]) Len() int {
	return uc.q.Len() + len(uc.ch)
}

func (uc *UnboundedChan[T]) Chan() <-chan T {
	return uc.ch
}

// Close the UnboundedChan. This mimics the behavior of closing the standard channel.
// **However**, there are some significant semnatic difference from the standard channel, if not handled properly
// will lead to goroutine and memory leaks
//
// tldr: if the data is nolonger needed after Close, consider using CloseAndDrain
//
// Similar:
//
// 1. Further write (Push) to UnboundedChan will panic;
// 2. Further read from the UnboundedChan drains the channel until zero value is returned.
// 3. Close Closed will panic
//
// Different:
//
// 1. UnboundedChan MUST be closed explicitly. Standard channel can get GC'd even if not closed, not this one unfortunately.
// Because of the internal implementation detail that uses goroutine to move data to channel, which will keepalive the objects.
//
// 2. reader need to make sure the UnboundedChan is drained after closing.
// Because of the internal implementation detail that uses goroutine to move data to channel,
// if the channel is not drained, that goroutine will be blocked and causes *goroutine leak* and *memory leak*
//
// 3. the standard channel can be drained with for-select (https://gobyexample.com/select) with default case to bail after the channel is closed.
// UnboundedChan cannot be drained reliably with that. This is because a background goroutine is feeding the internal channel from queue,
// there could still be items in queue when the channel is reading empty temporaily due to goroutine scheduling.
// To properly drain the UnboundedChan, keeps reading from channel until zero value is returned or until not ok
//
// Example:
//
//	uc.Close()
//	ch := uc.Chan()
//	for {
//	    v, ok := <- ch
//	    if !ok {
//	        break
//	    }
//	}
func (uc *UnboundedChan[T]) Close() {
	if uc.closed.Load() {
		log.Println("FATAL close of closed channel")
		panic("close of closed channel")
	}

	uc.q.Close()
	uc.closed.Store(true)
}

func (uc *UnboundedChan[T]) Closed() bool {
	return uc.closed.Load()
}

func (uc *UnboundedChan[T]) CloseAndDrain() {
	uc.Close()
	uc.Drain()
}

func (uc *UnboundedChan[T]) Drain() {
	for {
		_, _, err := uc.q.Pop()
		if err != nil {
			break
		}
	}

	for {
		_, ok := <-uc.ch
		if !ok {
			break
		}
	}
}