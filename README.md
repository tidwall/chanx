# chanx

[![GoDoc](https://img.shields.io/badge/api-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/tidwall/chanx)

A simple interface wrapper around a Go channel.

```go
// Make new channel. Provide a length to make a buffered channel.
func Make(length int) C

type C interface {
	// Send a messge to the channel. Returns false if the channel is closed.
	Send(v interface{}) (ok bool)
	// Recv a messge from the channel. Returns false if the channel is closed.
	Recv() (v interface{}, ok bool)
	// Close the channel. Returns false if the channel is already closed.
	Close() (ok bool)
	// Wait for the channel to close. Returns immediately if the channel is
	// already closed
	Wait()
}
```

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

chanx source code is available under the MIT [License](/LICENSE).

