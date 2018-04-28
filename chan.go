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
	waitg sync.WaitGroup // used for sleeping. gotta get our zzzs
	queue *nodeT         // items in the sender queue
	recvd *nodeT         // receive queue, receiver-only
	freed *nodeT         // freelist for minimizing allocation
	first *nodeT         // first item in recvd
	last  *nodeT         // last item in recvd
	count uintptr        // track the number of items in the queue
}

// Send sends a message of the receiver.
func (ch *Chan) Send(value interface{}) {
	var n *nodeT
	for {
		n = (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		if n == nil {
			n = new(nodeT)
			break
		}
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(n), unsafe.Pointer(n.next)) {
			break
		}
		runtime.Gosched()
	}
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
	atomic.AddUintptr(&ch.count, 1)
}

// Recv receives the next message.
func (ch *Chan) Recv() interface{} {
	for {
		if ch.recvd != nil {
			// new message, fist pump
			value := ch.recvd.value
			ch.recvd = ch.recvd.next
			if ch.recvd == nil {
				// add to received items to the free list
				for {
					freed := (*nodeT)(atomic.LoadPointer(
						(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
					))
					ch.last.next = freed
					if atomic.CompareAndSwapPointer(
						(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
						unsafe.Pointer(freed), unsafe.Pointer(ch.first)) {
						break
					}
					runtime.Gosched()
				}
			}
			atomic.AddUintptr(&ch.count, ^uintptr(0))
			return value
		}
		// let's load more messages from the sender queue.
		for {
			queue := (*nodeT)(atomic.LoadPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
			))
			if queue == nil {
				// sender queue is empty. put the receiver to sleep
				ch.waitg.Add(1)
				if atomic.CompareAndSwapPointer(
					(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
					unsafe.Pointer(queue), unsafe.Pointer(sleepN)) {
					ch.waitg.Wait()
				} else {
					ch.waitg.Done()
				}
			} else if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(queue), nil) {
				// we have an isolated queue of messages
				// reverse the queue
				var prev, next *nodeT
				var current = queue
				for current != nil {
					next = current.next
					current.next = prev
					prev = current
					current = next
				}
				// fill the recvd list
				ch.recvd = prev
				// set the first and last items for freeing later
				ch.first = prev
				ch.last = queue
				break
			}
			runtime.Gosched()
		}
	}
}

// Len returns the number of message in the sender queue.
func (ch *Chan) Len() int {
	return int(atomic.LoadUintptr(&ch.count))
}
