package main

import (
	"log"
	"runtime"
	"time"

	"github.com/fredwangwang/go-unboundedchannel"
)

func main() {
	for i := 0; i < 10; i++ {
		test()
	}

	runtime.GC()
	time.Sleep(1 * time.Second)

	log.Println("remaining goroutine", runtime.NumGoroutine())
	if runtime.NumGoroutine() >= 10 {
		panic("goroutine leaked")
	}
}

func test() {
	numInit := 100
	numPush := 1000
	numRead := 350

	// change this from NewUnboundedChanWithFinalizer to NewUnbounded will panic due to gorutine leak
	// uc := unboundedchannel.NewUnboundedChan[int](numInit)
	uc := unboundedchannel.NewUnboundedChanWithFinalizer[int](numInit)

	for i := 0; i < numPush; i++ {
		uc.Push(i)
	}

	ch := uc.Chan()
	for i := 0; i < numRead; i++ {
		<-ch
	}

	rem := uc.Len()
	log.Println("actual remaining", numPush-numRead, "reported remaining", rem, "upper bound remaining", rem+numInit)
}
