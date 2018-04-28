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

var sleepN = new(nodeT) // placeholder that indicates the receiver is sleeping

// nodeT is channel message
type nodeT struct {
	value interface{} // the message value. i hope it's a happy one
	next  *nodeT      // next item in the queue
}

// Chan represents a single-producer / single-consumer channel.
type Chan struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue *nodeT         // items in the sender queue
	recvd *nodeT         // receive queue, receiver-only
}

// Send sends a message of the receiver.
func (ch *Chan) Send(value interface{}) {
	n := new(nodeT)
	n.value = value
	var wake bool
	for {
		n.next = (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n.next == sleepN {
			// there's a sleep placeholder in the sender queue.
			// clear it and prepare to wake the receiver.
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n.next), unsafe.Pointer(n.next.next)) {
				wake = true
			}
		} else {
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n.next), unsafe.Pointer(n)) {
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
		// new message, fist pump
		value := ch.recvd.value
		ch.recvd = ch.recvd.next
		return value
	}
	// let's load more messages from the sender queue.
	var n *nodeT
	for {
		n = (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n == nil {
			// sender queue is empty. put the receiver to sleep
			ch.waitg.Add(1)
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n), unsafe.Pointer(sleepN)) {
				ch.waitg.Wait()
			} else {
				ch.waitg.Done()
			}
		} else if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
			unsafe.Pointer(n), nil) {
			break
		}
		runtime.Gosched()
	}
	// reverse queue
	var prev, next *nodeT
	var current = n
	for current != nil {
		next = current.next
		current.next = prev
		prev = current
		current = next
	}
	// set recvd list and return value
	value := prev.value
	ch.recvd = prev.next
	return value
}
