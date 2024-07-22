package main

import (
	"log"
	"runtime"
	"time"

	"github.com/fredwangwang/go-unboundedchannel"
)

func main() {
	for i := 0; i < 1000; i++ {
		test()
	}

	runtime.GC()
	time.Sleep(1 * time.Second)

	log.Println("remaining goroutine", runtime.NumGoroutine())
	if runtime.NumGoroutine() >= 1000 {
		panic("goroutine leaked")
	}
}

func test() {
	// change this from NewUnboundedChanWithFinalizer to NewUnbounded will panic due to gorutine leak
	// uc := unboundedchannel.NewUnboundedChan[int](100)
	uc := unboundedchannel.NewUnboundedChanWithFinalizer[int](100)

	for i := 0; i < 1000; i++ {
		uc.Push(i)
	}

	ch := uc.Chan()
	for i := 0; i < 200; i++ {
		<-ch
	}
}
