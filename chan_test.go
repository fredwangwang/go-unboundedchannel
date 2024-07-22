package unboundedchannel

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Single_Writer_Reader_Sequential_Ordering(t *testing.T) {
	uc := NewUnboundedChan[int](100)

	for i := 0; i < 10000; i++ {
		uc.Push(i)
	}

	ch := uc.Chan()
	for i := 0; i < 10000; i++ {
		v := <-ch
		assert.Equal(t, i, v)
	}
}

func Test_Single_Writer_Reader_Concurrent_Ordering(t *testing.T) {
	uc := NewUnboundedChan[int](100)

	wg := sync.WaitGroup{}
	wg.Add(2)

	var i int
	for i = 0; i < 200; i++ {
		uc.Push(i)
	}

	go func() {
		for {
			if i > 10000 {
				break
			}
			uc.Push(i)
			i++
			runtime.Gosched()
		}
		wg.Done()
	}()

	var res []int
	ch := uc.Chan()

	go func() {
		for i := 0; i < 10000; i++ {
			v := <-ch
			res = append(res, v)
			if i%20 == 1 {
				runtime.Gosched()
			}
		}
		wg.Done()
	}()

	wg.Wait()

	for i := 0; i < 10000; i++ {
		assert.Equal(t, i, res[i])
	}

	uc.CloseAndDrain()
}

func Test_Multi_Writer_Ordering(t *testing.T) {
	// each writer's item should be ordered

	uc := NewUnboundedChan[*int](100)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < 10000; i++ {
			i := i
			if i%2 == 0 {
				uc.Push(&i)
			}
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 10000; i++ {
			i := i
			if i%2 == 1 {
				uc.Push(&i)
			}
		}
		wg.Done()
	}()

	wg.Wait()
	uc.Close()

	ch := uc.Chan()
	res := []int{}
	for {
		v := <-ch
		if v == nil {
			break
		}

		res = append(res, *v)
	}

	assert.Equal(t, 10000, len(res))

	lastEven := -2
	lastOdd := -1
	for _, itm := range res {
		if itm%2 == 0 {
			assert.True(t, itm > lastEven, "%i should be larger than %i", itm, lastEven)
			lastEven = itm
		} else {
			assert.True(t, itm > lastOdd)
			lastOdd = itm
		}
	}
}

func Test_WriteTo_Closed_Should_Panice(t *testing.T) {
	uc := NewUnboundedChan[int](5)
	uc.Close()

	defer func() {
		if err := recover(); err == nil {
			t.Log("should panic")
			t.Fail()
		}
	}()
	uc.Push(1)
}

func Test_CloseAndDrain(t *testing.T) {
	uc := NewUnboundedChan[int](5)

	for i := 0; i < 1000; i++ {
		uc.Push(i)
	}

	// read something from the channel
	ch := uc.Chan()
	for i := 0; i < 20; i++ {
		<-ch
	}

	// drains the remaining
	uc.CloseAndDrain()

	assert.Equal(t, 0, len(uc.ch))
	assert.Equal(t, 0, uc.q.Len())
}
