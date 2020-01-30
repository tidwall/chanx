package chanx

type msg struct {
	v  interface{}
	ok bool
}

// C is a channel
type C interface {
	Send(v interface{}) (ok bool)
	Recv() (v interface{}, ok bool)
	Close() (ok bool)
}

// c ...
type c struct {
	c chan msg
}

// Make new channel. Provide a length to make a buffered channel.
func Make(length int) C {
	return &c{c: make(chan msg, length)}
}

// Send a messge to the channel. Returns false if the channel is closed or not
// initialized using Make().
func (c *c) Send(v interface{}) (ok bool) {
	defer func() { ok = recover() == nil }()
	c.c <- msg{v, true}
	return
}

// Recv a messge from the channel. Returns false if the channel is closed or
// not initialized using Make().
func (c *c) Recv() (v interface{}, ok bool) {
	select {
	case msg := <-c.c:
		return msg.v, msg.ok
	}
}

// Close the channel. Returns false if the channel is already closed or not
// initialized using Make().
func (c *c) Close() (ok bool) {
	defer func() { ok = recover() == nil }()
	close(c.c)
	return
}
