# `fastlane`

[![Build Status](https://img.shields.io/travis/tidwall/fastlane.svg?style=flat-square)](https://travis-ci.org/tidwall/fastlane)
[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/tidwall/fastlane)


Fast single-producer / single-consumer channels for Go.

A fastlane channel works similar to a standard Go channel with the following exceptions:

- It does not have a close method. A sender must send the receiver a custom close message.
- It's unbounded and has no buffering. There's lock-free linked list under the hood.
- It expects to be communicating over a maximum of two goroutines. One for `Send` and one for `Recv`. 

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

Here's a simple wrapper for creating a channel for sending a custom types as a pointer:

```go
type MyType struct {
	Hiya int
}

type ChanMyType struct{ base fastlane.ChanPointer }

func (ch *ChanMyType) Send(value *MyType) {
	ch.base.Send(unsafe.Pointer(value))
}
func (ch *ChanMyType) Recv() *MyType {
	return (*MyType)(ch.base.Recv())
}
```


## Performance

The benchmark tests the speed of sending integers between two goroutines.

```
$ go test -run none -bench .
``` 

MacBook Pro 15" 2.8 GHz Intel Core i7 (darwin/amd64)

```
BenchmarkFastlaneChan-8       	30000000	        40.4 ns/op
BenchmarkGoChan100-8          	20000000	        68.9 ns/op
BenchmarkGoChan10-8           	20000000	        74.4 ns/op
BenchmarkGoChanUnbuffered-8   	10000000	       197 ns/op
```

Mac mini i7-3615QM CPU @ 2.30GHz (linux/amd64)

```
BenchmarkFastlaneChan-8       	20000000	        67.3 ns/op
BenchmarkGoChan100-8          	10000000	       181 ns/op
BenchmarkGoChan10-8           	 5000000	       223 ns/op
BenchmarkGoChanUnbuffered-8   	 5000000	       595 ns/op
```

Raspberry Pi 3 (linux/arm64)

```
BenchmarkFastlaneChan-4       	10000000	       213 ns/op
BenchmarkGoChan100-4          	 3000000	       406 ns/op
BenchmarkGoChan10-4           	 3000000	       578 ns/op
BenchmarkGoChanUnbuffered-4   	 1000000	      1405 ns/op
```

Raspberry Pi 3 (linux/arm)

```
BenchmarkFastlaneChan-4       	 5000000	       334 ns/op
BenchmarkGoChan100-4          	 2000000	       669 ns/op
BenchmarkGoChan10-4           	 2000000	       936 ns/op
BenchmarkGoChanUnbuffered-4   	 1000000	      2370 ns/op
```

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

`fastlane` source code is available under the MIT [License](/LICENSE).

