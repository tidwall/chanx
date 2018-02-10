package fastlane

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// sleep is a placeholder that indicates the receiver is sleeping
var sleep = new(node)

// node is channel message
type node struct {
	valu interface{} // the message value. i hope it's a happy one
	prev *node       // used by the receiver for tracking backwards
	next *node       // next item in the queue or freelist
}

// Chan represents a single-producer / single-consumer channel.
type Chan struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue *node          // sender queue, sender -> receiver
	recvd *node          // receive queue, receiver-only
	freed *node          // freed queue, receiver -> sender
	avail *node          // avail items, sender-only
}

func (ch *Chan) load() *node {
	return (*node)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
	))
}

func (ch *Chan) cas(old, new *node) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		unsafe.Pointer(old), unsafe.Pointer(new))
}

func (ch *Chan) new() *node {
	if ch.avail != nil {
		n := ch.avail
		ch.avail = ch.avail.next
		return n
	}
	for {
		ch.avail = (*node)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		if ch.avail == nil {
			return &node{}
		}
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(ch.avail), nil) {
			return ch.new()
		}
		runtime.Gosched()
	}
}

func (ch *Chan) free(recvd *node) {
	for {
		freed := (*node)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(freed), unsafe.Pointer(recvd)) {
			return
		}
		runtime.Gosched()
	}
}

// Send sends a messages to the receiver.
func (ch *Chan) Send(value interface{}) {
	n := ch.new()
	n.valu = value
	var wake bool
	for {
		n.next = ch.load()
		if n.next == sleep {
			// there's a sleep placeholder in the sender queue.
			// clear it and prepare to wake the receiver.
			if ch.cas(n.next, n.next.next) {
				wake = true
			}
		} else {
			if ch.cas(n.next, n) {
				break
			}
		}
		runtime.Gosched()
	}
	if wake {
		// wake up the receiver
		ch.waitg.Done()
	}
}

// Recv receives the next message.
func (ch *Chan) Recv() interface{} {
	if ch.recvd != nil {
		// we have a received item
		v := ch.recvd.valu
		// set the value to zero allowing it to be freed by the GC.
		(*(*[2]uintptr)(unsafe.Pointer(&ch.recvd.valu)))[1] = 0
		if ch.recvd.prev == nil {
			// we're at the end of the recieve queue. put the available
			// nodes into the freelist.
			ch.free(ch.recvd)
			ch.recvd = nil
		} else {
			ch.recvd = ch.recvd.prev
			ch.recvd.next.prev = nil
		}
		return v
	}
	// no received items, let's load more from the sender queue.
	var n *node
	for {
		n = ch.load()
		if n == nil {
			// empty sender queue. put the receiver to sleep
			ch.waitg.Add(1)
			if ch.cas(n, sleep) {
				ch.waitg.Wait()
			} else {
				ch.waitg.Done()
			}
		} else if ch.cas(n, nil) {
			break
		}
		runtime.Gosched()
	}
	// set the prev pointers for tracking backwards
	for {
		if n.next == nil {
			break
		}
		n.next.prev = n
		n = n.next
	}
	// fill the receive queue
	ch.recvd = n
	return ch.Recv()
}
