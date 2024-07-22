# go-unboundedchannel
An elastically grow channel impl for golang, used in cases where sudden backpressure cannot be afford to be blocked or dropped.

## Advantage over a large buffered channel
buffered channel actually consumes memory when created. In cases where you anticipate mostly only small buffer requrired, but need to account for sudden expansion in magnitude,
it would be wasteful to allocate all those buffers up front. Instead, unbounded channel will start with a initial buffer that suits most of the cases, and elastically grow if needed.

## Disadvantage over buffered channel

The ealstically growing feature is powered by a background goroutine. Due to the lack of `weakref`,
the goroutine unfortunatelly holds reference to the struct itself, preventing it to be automatically
destroyed through finalizer. Thus it needs some special attention when using. Check the comment on
`chan.go:UnboundedChan.Close` for more info.

## TODO

maybe looking into finalizer and weakref again to see if the disadvantage can be improved.