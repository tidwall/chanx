# `fastlane`

[![Build Status](https://img.shields.io/travis/tidwall/fastlane.svg?style=flat-square)](https://travis-ci.org/tidwall/fastlane)
[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/tidwall/fastlane)


Fast multi-producer / single-consumer channels for Go.

A fastlane channel works similar to a standard Go channel with the following exceptions:

- It does not have a close method. A sender must send the receiver a custom close message.
- It's unbounded and has no buffering. Sending a message is a non-blocking O(1) operation.
- There can be multiple goroutines that call `Send`, but only one for `Recv`.

Just like standard Go channels, a fastlane channel guarantees order preservation and has no data loss.

# Getting Started

### Installing

To start using fastlane, install Go and run `go get`:

```sh
$ go get -u github.com/tidwall/fastlane
```

This will retrieve the library.

### Usage

There're only two functions `Send` and `Recv`.

```go
// chan.go
package main

import "github.com/tidwall/fastlane"

func main() {
	var ch fastlane.Chan

	go func() { ch.Send("ping") }()

	v := ch.Recv()
	println(v.(string))
}
```

```sh
$ go run chan.go 
ping
```


## Channel types

There's currently three types of channels, `Chan` for `interface{}`, `ChanUint64` for `uint64`, and `ChanPointer` for `unsafe.Pointer`.

The `ChanUint64` and `ChanPointer` often perform better than the generic `Chan` and should be used when possible.

Here's a simple wrapper for creating a channel for sending a custom type as a pointer:

```go
type MyType struct {
	Hiya string
}

type MyChan struct{ base fastlane.ChanPointer }

func (ch *MyChan) Send(value *MyType) {
	ch.base.Send(unsafe.Pointer(value))
}
func (ch *MyChan) Recv() *MyType {
	return (*MyType)(ch.base.Recv())
}
```

```go
var ch MyChan

go func() { ch.Send(&MyType{Hiya: "howdy!"}) }()

v := ch.Recv()
println(v.Hiya)
```

## Performance

The benchmark tests the speed of consuming integers from 100 producers.  
*go version go1.10.1*

```
$ go test -run none -bench .
``` 

MacBook Pro 15" 2.8 GHz Intel Core i7 (darwin/amd64)

```
Benchmark100ProducerFastlaneChan-8       	20000000	        85.9 ns/op
Benchmark100ProducerGoChan100-8          	 3000000	       569 ns/op
Benchmark100ProducerGoChan10-8           	 2000000	       728 ns/op
Benchmark100ProducerGoChanUnbuffered-8   	 2000000	       686 ns/op
```

Mac mini i7-3615QM CPU @ 2.30GHz (linux/amd64)

```
Benchmark100ProducerFastlaneChan-8       	20000000	       101 ns/op
Benchmark100ProducerGoChan100-8          	 3000000	       424 ns/op
Benchmark100ProducerGoChan10-8           	 2000000	       586 ns/op
Benchmark100ProducerGoChanUnbuffered-8   	 2000000	       790 ns/op
```

Raspberry Pi 3 (linux/arm64)

```
Benchmark100ProducerFastlaneChan-4       	10000000	       179 ns/op
Benchmark100ProducerGoChan100-4          	 2000000	       976 ns/op
Benchmark100ProducerGoChan10-4           	 1000000	      1043 ns/op
Benchmark100ProducerGoChanUnbuffered-4   	 1000000	      1093 ns/op
```

Raspberry Pi 3 (linux/arm)

```
Benchmark100ProducerFastlaneChan-4       	 5000000	       279 ns/op
Benchmark100ProducerGoChan100-4          	 1000000	      1345 ns/op
Benchmark100ProducerGoChan10-4           	 1000000	      1371 ns/op
Benchmark100ProducerGoChanUnbuffered-4   	 1000000	      1635 ns/op
```

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

`fastlane` source code is available under the MIT [License](/LICENSE).

