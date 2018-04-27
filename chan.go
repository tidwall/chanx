// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package fastlane

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

//go:generate cp gen.go /tmp/chan_gen.go
//go:generate go run /tmp/chan_gen.go
//go:generate go fmt

var sleepN = &nodeT{} // placeholder that indicates the receiver is sleeping
var emptyN = &nodeT{} // placeholder that indicates the ready list is empty

// nodeT is channel message
type nodeT struct {
	value interface{}
	next  *nodeT
}

// Chan ...
type Chan struct {
	waitg  sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue  *nodeT         // items in the sender queue
	readys *nodeT         // items ready for receiving
}

func (ch *Chan) load() *nodeT {
	return (*nodeT)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
	))
}

func (ch *Chan) cas(old, new *nodeT) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		unsafe.Pointer(old), unsafe.Pointer(new))
}

// Send sends a message of the receiver.
func (ch *Chan) Send(value interface{}) {
	n := &nodeT{value: value}
	var wake bool
	for {
		n.next = ch.load()
		if n.next == sleepN {
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
		ch.waitg.Done()
	}
}

// Recv receives the next message.
func (ch *Chan) Recv() interface{} {
	// look in receiver list for items before checking the sender queue.
	for {
		readys := (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
		))
		if readys == nil || readys == emptyN {
			break
		}
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
			unsafe.Pointer(readys), unsafe.Pointer(readys.next)) {
			return readys.value
		}
		runtime.Gosched()
	}

	// let's load more messages from the sender queue.
	var queue *nodeT
	for {
		queue = ch.load()
		if queue == nil {
			// sender queue is empty. put the receiver to sleep
			ch.waitg.Add(1)
			if ch.cas(queue, sleepN) {
				ch.waitg.Wait()
			} else {
				ch.waitg.Done()
			}
		} else if ch.cas(queue, nil) {
			// empty the queue
			break
		}
		runtime.Gosched()
	}
	// reverse the order
	var prev *nodeT
	var current = queue
	var next *nodeT
	for current != nil {
		next = current.next
		current.next = prev
		prev = current
		current = next
	}
	if prev.next != nil {
		// we have ordered items that must be handled later
		for {
			readys := (*nodeT)(atomic.LoadPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
			))
			if readys != emptyN {
				queue.next = readys
			} else {
				queue.next = nil
			}
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
				unsafe.Pointer(readys), unsafe.Pointer(prev.next)) {
				break
			}
			runtime.Gosched()
		}
	}
	return prev.value
}

// // sleepN is a placeholder that indicates the receiver is sleeping
// var sleepN = &nodeT{}

// // nodeT is channel message
// type nodeT struct {
// 	valu interface{} // the message value. i hope it's a happy one
// 	prev *nodeT      // used by the receiver for tracking backwards
// 	next *nodeT      // next item in the queue or freelist
// }

// // Chan represents a single-producer / single-consumer channel.
// type Chan struct {
// 	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
// 	queue *nodeT         // sender queue, sender -> receiver
// 	recvd *nodeT         // receive queue, receiver-only
// 	rtail *nodeT         // tail of receive queue, receiver-only
// 	freed *nodeT         // freed queue, receiver -> sender
// 	avail *nodeT         // avail items, sender-only
// }

// func (ch *Chan) load() *nodeT {
// 	return (*nodeT)(atomic.LoadPointer(
// 		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
// 	))
// }

// func (ch *Chan) cas(old, new *nodeT) bool {
// 	return atomic.CompareAndSwapPointer(
// 		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
// 		unsafe.Pointer(old), unsafe.Pointer(new))
// }

// func (ch *Chan) new() *nodeT {
// 	if ch.avail != nil {
// 		n := ch.avail
// 		ch.avail = ch.avail.next
// 		return n
// 	}
// 	for {
// 		ch.avail = (*nodeT)(atomic.LoadPointer(
// 			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
// 		))
// 		if ch.avail == nil {
// 			return &nodeT{}
// 		}
// 		if atomic.CompareAndSwapPointer(
// 			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
// 			unsafe.Pointer(ch.avail), nil) {
// 			return ch.new()
// 		}
// 		runtime.Gosched()
// 	}
// }

// func (ch *Chan) free(recvd *nodeT) {
// 	for {
// 		freed := (*nodeT)(atomic.LoadPointer(
// 			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
// 		))
// 		ch.rtail.next = freed
// 		if atomic.CompareAndSwapPointer(
// 			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
// 			unsafe.Pointer(freed), unsafe.Pointer(recvd)) {
// 			return
// 		}
// 		runtime.Gosched()
// 	}
// }

// // Send sends a message to the receiver.
// func (ch *Chan) Send(value interface{}) {
// 	n := ch.new()
// 	n.valu = value
// 	var wake bool
// 	for {
// 		n.next = ch.load()
// 		if n.next == sleepN {
// 			// there's a sleep placeholder in the sender queue.
// 			// clear it and prepare to wake the receiver.
// 			if ch.cas(n.next, n.next.next) {
// 				wake = true
// 			}
// 		} else {
// 			if ch.cas(n.next, n) {
// 				break
// 			}
// 		}
// 		runtime.Gosched()
// 	}
// 	if wake {
// 		// wake up the receiver
// 		ch.waitg.Done()
// 	}
// }

// // Recv receives the next message.
// func (ch *Chan) Recv() interface{} {
// 	if ch.recvd != nil {
// 		// new message, fist pump
// 		v := ch.recvd.valu
// 		if ch.recvd.prev == nil {
// 			// we're at the end of the recieve queue. put the available
// 			// nodes into the freelist.
// 			ch.free(ch.recvd)
// 			ch.recvd = nil
// 		} else {
// 			ch.recvd = ch.recvd.prev
// 			ch.recvd.next.prev = nil
// 		}
// 		return v
// 	}
// 	// let's load more messages from the sender queue.
// 	var n *nodeT
// 	for {
// 		n = ch.load()
// 		if n == nil {
// 			// sender queue is empty. put the receiver to sleep
// 			ch.waitg.Add(1)
// 			if ch.cas(n, sleepN) {
// 				ch.waitg.Wait()
// 			} else {
// 				ch.waitg.Done()
// 			}
// 		} else if ch.cas(n, nil) {
// 			break
// 		}
// 		runtime.Gosched()
// 	}
// 	// set the prev pointers for tracking backwards
// 	for n.next != nil {
// 		n.next.prev = n
// 		n = n.next
// 	}
// 	ch.recvd = n // fill receive queue
// 	ch.rtail = n // mark the free tail
// 	return ch.Recv()
// }
