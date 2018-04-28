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

var sleepUint64 = new(nodeUint64) // placeholder that indicates the receiver is sleeping

// nodeUint64 is channel message
type nodeUint64 struct {
	value uint64      // the message value. i hope it's a happy one
	next  *nodeUint64 // next item in the queue
}

// ChanUint64 represents a single-producer / single-consumer channel.
type ChanUint64 struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue *nodeUint64    // items in the sender queue
	recvd *nodeUint64    // receive queue, receiver-only
}

// Send sends a message of the receiver.
func (ch *ChanUint64) Send(value uint64) {
	n := new(nodeUint64)
	n.value = value
	var wake bool
	for {
		n.next = (*nodeUint64)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n.next == sleepUint64 {
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
func (ch *ChanUint64) Recv() uint64 {
	if ch.recvd != nil {
		// new message, fist pump
		value := ch.recvd.value
		ch.recvd = ch.recvd.next
		return value
	}
	// let's load more messages from the sender queue.
	var n *nodeUint64
	for {
		n = (*nodeUint64)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n == nil {
			// sender queue is empty. put the receiver to sleep
			ch.waitg.Add(1)
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n), unsafe.Pointer(sleepUint64)) {
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
	var prev, next *nodeUint64
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

var sleepPointer = new(nodePointer) // placeholder that indicates the receiver is sleeping

// nodePointer is channel message
type nodePointer struct {
	value unsafe.Pointer // the message value. i hope it's a happy one
	next  *nodePointer   // next item in the queue
}

// ChanPointer represents a single-producer / single-consumer channel.
type ChanPointer struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue *nodePointer   // items in the sender queue
	recvd *nodePointer   // receive queue, receiver-only
}

// Send sends a message of the receiver.
func (ch *ChanPointer) Send(value unsafe.Pointer) {
	n := new(nodePointer)
	n.value = value
	var wake bool
	for {
		n.next = (*nodePointer)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n.next == sleepPointer {
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
func (ch *ChanPointer) Recv() unsafe.Pointer {
	if ch.recvd != nil {
		// new message, fist pump
		value := ch.recvd.value
		ch.recvd = ch.recvd.next
		return value
	}
	// let's load more messages from the sender queue.
	var n *nodePointer
	for {
		n = (*nodePointer)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n == nil {
			// sender queue is empty. put the receiver to sleep
			ch.waitg.Add(1)
			if atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
				unsafe.Pointer(n), unsafe.Pointer(sleepPointer)) {
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
	var prev, next *nodePointer
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
