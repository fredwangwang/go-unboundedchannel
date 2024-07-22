# go-unboundedchannel
An elastically grow channel impl for golang, used in cases where sudden backpressure cannot be afford to be blocked or dropped.

## Advantage over a large buffered channel
buffered channel actually consumes memory when created. In cases where you anticipate mostly only small buffer requrired, but need to account for sudden expansion in magnitude,
it would be wasteful to allocate all those buffers up front. Instead, unbounded channel will start with a initial buffer that suits most of the cases, and elastically grow if needed.

## Disadvantage over buffered channel

The ealstically growing feature is powered by a background goroutine, which means if the unbounded channel is not closed and drained properly, it could cause goroutine and memory leak.
This can be remediated using `NewUnboundedChanWithFinalizer` to automatically cleanup the channel if no external refernce to the channel exists.
