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

## Performance

The benchmark tests the speed of sending integers between two goroutines.

```
$ go test -run none -bench .
``` 

MacBook Pro 15" 2.8 GHz Intel Core i7 (darwin/amd64)

```
BenchmarkFastlaneChan-8   	20000000	        72.7 ns/op
BenchmarkGoChan-8         	 5000000	       241 ns/op
```

Mac mini i7-3615QM CPU @ 2.30GHz (linux/amd64)

```
BenchmarkFastlaneChan-8   	10000000	       105 ns/op
BenchmarkGoChan-8         	 3000000	       602 ns/op
```

Raspberry Pi 3 (linux/arm64)

```
BenchmarkFastlaneChan-4   	 5000000	       393 ns/op
BenchmarkGoChan-4         	 1000000	      1337 ns/op
```

Raspberry Pi 3 (linux/arm)

```
BenchmarkFastlaneChan-4   	 3000000	       535 ns/op
BenchmarkGoChan-4         	 1000000	      2195 ns/op
```

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

`fastlane` source code is available under the MIT [License](/LICENSE).

