package chanx

import "sync"

type msg struct {
	v  interface{}
	ok bool
}

// C is a channel
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

type c struct {
	mu     sync.Mutex
	cond   *sync.Cond
	c      chan msg
	closed bool
}

// Make new channel. Provide a length to make a buffered channel.
func Make(length int) C {
	c := &c{c: make(chan msg, length)}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *c) Send(v interface{}) (ok bool) {
	defer func() { ok = recover() == nil }()
	c.c <- msg{v, true}
	return
}

func (c *c) Recv() (v interface{}, ok bool) {
	select {
	case msg := <-c.c:
		return msg.v, msg.ok
	}
}

func (c *c) Close() (ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() { ok = recover() == nil }()
	close(c.c)
	c.closed = true
	c.cond.Broadcast()
	return
}

func (c *c) Wait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for {
		if c.closed {
			return
		}
		c.cond.Wait()
	}
}
