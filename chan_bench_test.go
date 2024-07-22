package unboundedchannel

import (
	"sync"
	"testing"
)

type PFunc func(int)

func Benchmark_Unbuffered_Chan(b *testing.B) {
	ch := make(chan int)
	pfunc := func(i int) {
		ch <- i
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)
		go producer(pfunc, &wg)
		consumer(b, ch, &wg)
		// wg.Wait()
	}
}

func Benchmark_Buffered_10_Chan(b *testing.B) {
	ch := make(chan int, 10)
	pfunc := func(i int) {
		ch <- i
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)
		go producer(pfunc, &wg)
		consumer(b, ch, &wg)
		// wg.Wait()
	}
}

func Benchmark_Buffered_100_Chan(b *testing.B) {
	ch := make(chan int, 100)
	pfunc := func(i int) {
		ch <- i
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)

		go producer(pfunc, &wg)
		consumer(b, ch, &wg)
		// wg.Wait()
	}
}

func Benchmark_Buffered_1000_Chan(b *testing.B) {
	ch := make(chan int, 1000)
	pfunc := func(i int) {
		ch <- i
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)

		go producer(pfunc, &wg)
		consumer(b, ch, &wg)
		// wg.Wait()
	}
}

func Benchmark_Unbounded_10_Chan(b *testing.B) {
	ch := NewUnboundedChan[int](10)
	pfunc := func(i int) {
		ch.Push(i)
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)

		go producer(pfunc, &wg)
		consumer(b, ch.Chan(), &wg)
		// wg.Wait()
	}
}

func Benchmark_Unbounded_100_Chan(b *testing.B) {
	ch := NewUnboundedChan[int](100)
	pfunc := func(i int) {
		ch.Push(i)
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)

		go producer(pfunc, &wg)
		consumer(b, ch.Chan(), &wg)
		// wg.Wait()
	}
}

func Benchmark_Unbounded_1000_Chan(b *testing.B) {
	ch := NewUnboundedChan[int](1000)
	pfunc := func(i int) {
		ch.Push(i)
	}

	wg := sync.WaitGroup{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// wg.Add(2)

		go producer(pfunc, &wg)
		consumer(b, ch.Chan(), &wg)
		// wg.Wait()
	}
}

// func BenchmarkJonbodnerUnboundedChan(b *testing.B) {
// 	s, r := unbounded.MakeInfinite()

// 	pfunc := func(i int) {
// 		s <- i
// 	}

// 	wg := sync.WaitGroup{}

// 	for n := 0; n < b.N; n++ {
// 		// wg.Add(2)

// 		go producer(pfunc, &wg)
// 		consumerInterface(b, r, &wg)
// 		// wg.Wait()
// 	}
// }

func producer(p PFunc, wg *sync.WaitGroup) {
	for i := 0; i < 1000; i++ {
		p(i)
	}
	// log.Println("producer done")
	// wg.Done()
}

func consumer(b *testing.B, c <-chan int, wg *sync.WaitGroup) {
	for i := 0; i < 1000; i++ {
		<-c
	}
	// log.Println("consumer done")
	// wg.Done()
}

// func consumerInterface(b *testing.B, c <-chan interface{}, wg *sync.WaitGroup) {
// 	for i := 0; i < 10000; i++ {
// 		<-c
// 	}

// 	select {
// 	case <-c:
// 		b.Fail()
// 	default:
// 	}

// 	// wg.Done()
// }
