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

// sleepUint64 is a placeholder that indicates the receiver is sleeping
var sleepUint64 = &nodeUint64{}

// nodeUint64 is channel message
type nodeUint64 struct {
	valu uint64      // the message value. i hope it's a happy one
	prev *nodeUint64 // used by the receiver for tracking backwards
	next *nodeUint64 // next item in the queue or freelist
}

// ChanUint64 represents a single-producer / single-consumer channel.
type ChanUint64 struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue *nodeUint64    // sender queue, sender -> receiver
	recvd *nodeUint64    // receive queue, receiver-only
	freed *nodeUint64    // freed queue, receiver -> sender
	avail *nodeUint64    // avail items, sender-only
}

func (ch *ChanUint64) load() *nodeUint64 {
	return (*nodeUint64)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
	))
}

func (ch *ChanUint64) cas(old, new *nodeUint64) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		unsafe.Pointer(old), unsafe.Pointer(new))
}

func (ch *ChanUint64) new() *nodeUint64 {
	if ch.avail != nil {
		n := ch.avail
		ch.avail = ch.avail.next
		return n
	}
	for {
		ch.avail = (*nodeUint64)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		if ch.avail == nil {
			return &nodeUint64{}
		}
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(ch.avail), nil) {
			return ch.new()
		}
		runtime.Gosched()
	}
}

func (ch *ChanUint64) free(recvd *nodeUint64) {
	for {
		freed := (*nodeUint64)(atomic.LoadPointer(
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

// Send sends a message to the receiver.
func (ch *ChanUint64) Send(value uint64) {
	n := ch.new()
	n.valu = value
	var wake bool
	for {
		n.next = ch.load()
		if n.next == sleepUint64 {
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
func (ch *ChanUint64) Recv() uint64 {
	if ch.recvd != nil {
		// we have a received item
		v := ch.recvd.valu
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
	var n *nodeUint64
	for {
		n = ch.load()
		if n == nil {
			// empty sender queue. put the receiver to sleep
			ch.waitg.Add(1)
			if ch.cas(n, sleepUint64) {
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
	for n.next != nil {
		n.next.prev = n
		n = n.next
	}
	// fill the receive queue
	ch.recvd = n
	return ch.Recv()
}

// sleepPointer is a placeholder that indicates the receiver is sleeping
var sleepPointer = &nodePointer{}

// nodePointer is channel message
type nodePointer struct {
	valu unsafe.Pointer // the message value. i hope it's a happy one
	prev *nodePointer   // used by the receiver for tracking backwards
	next *nodePointer   // next item in the queue or freelist
}

// ChanPointer represents a single-producer / single-consumer channel.
type ChanPointer struct {
	waitg sync.WaitGroup // used for sleeping. gotta get our zzz's
	queue *nodePointer   // sender queue, sender -> receiver
	recvd *nodePointer   // receive queue, receiver-only
	freed *nodePointer   // freed queue, receiver -> sender
	avail *nodePointer   // avail items, sender-only
}

func (ch *ChanPointer) load() *nodePointer {
	return (*nodePointer)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
	))
}

func (ch *ChanPointer) cas(old, new *nodePointer) bool {
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		unsafe.Pointer(old), unsafe.Pointer(new))
}

func (ch *ChanPointer) new() *nodePointer {
	if ch.avail != nil {
		n := ch.avail
		ch.avail = ch.avail.next
		return n
	}
	for {
		ch.avail = (*nodePointer)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		if ch.avail == nil {
			return &nodePointer{}
		}
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(ch.avail), nil) {
			return ch.new()
		}
		runtime.Gosched()
	}
}

func (ch *ChanPointer) free(recvd *nodePointer) {
	for {
		freed := (*nodePointer)(atomic.LoadPointer(
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

// Send sends a message to the receiver.
func (ch *ChanPointer) Send(value unsafe.Pointer) {
	n := ch.new()
	n.valu = value
	var wake bool
	for {
		n.next = ch.load()
		if n.next == sleepPointer {
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
func (ch *ChanPointer) Recv() unsafe.Pointer {
	if ch.recvd != nil {
		// we have a received item
		v := ch.recvd.valu
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
	var n *nodePointer
	for {
		n = ch.load()
		if n == nil {
			// empty sender queue. put the receiver to sleep
			ch.waitg.Add(1)
			if ch.cas(n, sleepPointer) {
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
	for n.next != nil {
		n.next.prev = n
		n = n.next
	}
	// fill the receive queue
	ch.recvd = n
	return ch.Recv()
}
