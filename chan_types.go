// CODE GENERATED; DO NOT EDIT

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

// nodeUint64 is channel message
type nodeUint64 struct {
	value uint64      // the message value. i hope it's a happy one
	next  *nodeUint64 // next item in the queue
}

// ChanUint64 represents a single-producer / single-consumer channel.
type ChanUint64 struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzzs
	queue *nodeUint64    // items in the sender queue
	recvd *nodeUint64    // receive queue, receiver-only
	sleep nodeUint64     // resuable indicates the receiver is sleeping
}

// Send sends a message of the receiver.
func (ch *ChanUint64) Send(value uint64) {
	n := &nodeUint64{value: value}
	var wake bool
	for {
		n.next = (*nodeUint64)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n.next == &ch.sleep {
			// there's a sleep placeholder in the sender queue.
			// clear it and prepare to wake the receiver.
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n.next), unsafe.Pointer(n.next.next)) {
				// wake up the receiver
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
		ch.waitg.Done()
	}
}

// Recv receives the next message.
func (ch *ChanUint64) Recv() uint64 {
	for {
		if ch.recvd != nil {
			// new message, fist pump
			value := ch.recvd.value
			ch.recvd = ch.recvd.next
			return value
		}
		// let's load more messages from the sender queue.
		for {
			queue := (*nodeUint64)(atomic.LoadPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
			))
			if queue == nil {
				// sender queue is empty. put the receiver to sleep
				ch.waitg.Add(1)
				if atomic.CompareAndSwapPointer(
					(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
					unsafe.Pointer(queue), unsafe.Pointer(&ch.sleep)) {
					ch.waitg.Wait()
				} else {
					ch.waitg.Done()
				}
			} else if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(queue), nil) {
				// we have an isolated queue of messages
				// reverse the queue and fill the recvd list
				for queue != nil {
					next := queue.next
					queue.next = ch.recvd
					ch.recvd = queue
					queue = next
				}
				break
			}
			runtime.Gosched()
		}
	}
}

// nodePointer is channel message
type nodePointer struct {
	value unsafe.Pointer // the message value. i hope it's a happy one
	next  *nodePointer   // next item in the queue
}

// ChanPointer represents a single-producer / single-consumer channel.
type ChanPointer struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzzs
	queue *nodePointer   // items in the sender queue
	recvd *nodePointer   // receive queue, receiver-only
	sleep nodePointer    // resuable indicates the receiver is sleeping
}

// Send sends a message of the receiver.
func (ch *ChanPointer) Send(value unsafe.Pointer) {
	n := &nodePointer{value: value}
	var wake bool
	for {
		n.next = (*nodePointer)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n.next == &ch.sleep {
			// there's a sleep placeholder in the sender queue.
			// clear it and prepare to wake the receiver.
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n.next), unsafe.Pointer(n.next.next)) {
				// wake up the receiver
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
		ch.waitg.Done()
	}
}

// Recv receives the next message.
func (ch *ChanPointer) Recv() unsafe.Pointer {
	for {
		if ch.recvd != nil {
			// new message, fist pump
			value := ch.recvd.value
			ch.recvd = ch.recvd.next
			return value
		}
		// let's load more messages from the sender queue.
		for {
			queue := (*nodePointer)(atomic.LoadPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
			))
			if queue == nil {
				// sender queue is empty. put the receiver to sleep
				ch.waitg.Add(1)
				if atomic.CompareAndSwapPointer(
					(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
					unsafe.Pointer(queue), unsafe.Pointer(&ch.sleep)) {
					ch.waitg.Wait()
				} else {
					ch.waitg.Done()
				}
			} else if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(queue), nil) {
				// we have an isolated queue of messages
				// reverse the queue and fill the recvd list
				for queue != nil {
					next := queue.next
					queue.next = ch.recvd
					ch.recvd = queue
					queue = next
				}
				break
			}
			runtime.Gosched()
		}
	}
}
