package chan2

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

const (
	opened = iota
	closed
	sleeping
)

type nodeT struct {
	value int
	prev  *nodeT
	next  *nodeT
}

// Chan ...
type Chan struct {
	locker uintptr
	queue  *nodeT
	recvd  *nodeT // receive queue, receiver-only
	rtail  *nodeT // tail of receive queue, receiver-only
	freed  *nodeT
	avail  *nodeT

	// wg     sync.WaitGroup
	// ready  []int
	// locker uintptr
	// status int
	// queue  []int
}

// Send ...
func (ch *Chan) Send(value int) {
	n := ch.new()
	n.value = value
	for {
		queue := (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		n.next = queue
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
			unsafe.Pointer(queue), unsafe.Pointer(n)) {
			break
		}
		runtime.Gosched()
	}
}

func (ch *Chan) free(recvd *nodeT) {
	for {
		freed := (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		ch.rtail.next = freed
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(freed), unsafe.Pointer(recvd)) {
			return
		}
		runtime.Gosched()
	}
}

func (ch *Chan) new() *nodeT {
	if ch.avail != nil {
		n := ch.avail
		ch.avail = ch.avail.next
		return n
	}
	for {
		ch.avail = (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
		))
		if ch.avail == nil {
			return &nodeT{}
		}
		if atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.freed)),
			unsafe.Pointer(ch.avail), nil) {
			return ch.new()
		}
		runtime.Gosched()
	}
}

// Recv ...
func (ch *Chan) Recv() int {
	if ch.recvd != nil {
		// new message, fist pump
		v := ch.recvd.value
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

	// if ch.ready != nil {
	// 	n := ch.ready
	// 	ch.ready = ch.ready.next
	// 	ch.free(n)
	// 	return n.value
	// }
	var n *nodeT
	for {
		n = (*nodeT)(atomic.LoadPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
		))
		if n != nil && atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&ch.queue)),
			unsafe.Pointer(n), nil) {
			break
		}
		runtime.Gosched()
	}
	// // reverse the order
	// var prev, next *nodeT
	// var current = queue
	// for current != nil {
	// 	next = current.next
	// 	current.next = prev
	// 	prev = current
	// 	current = next
	// }
	// ch.ready = prev
	for n.next != nil {
		n.next.prev = n
		n = n.next
	}
	ch.recvd = n // fill receive queue
	ch.rtail = n // mark the free tail

	return ch.Recv()
}

// for {
// 	ch.lock()
// 	if len(ch.queue) == 0 {
// 		ch.unlock()
// 		continue
// 	}
// 	value := ch.queue[0]
// 	ch.queue = ch.queue[1:]
// 	ch.unlock()
// 	return value
// 	// if len(ch.ready) > 0 {
// 	// 	value := ch.ready[0]
// 	// 	ch.ready = ch.ready[1:]
// 	// 	return value
// 	// }
// 	// for {
// 	// 	ch.lock()
// 	// 	if len(ch.queue) == 0 {
// 	// 		ch.status = sleeping
// 	// 		ch.unlock()
// 	// 		ch.wg.Add(1)
// 	// 		ch.wg.Wait()
// 	// 		continue
// 	// 	}
// 	// 	ch.ready = ch.queue
// 	// 	ch.unlock()
// 	// 	break
// 	// }
// }

// var sleepN = &nodeT{} // placeholder that indicates the receiver is sleeping
// var emptyN = &nodeT{} // placeholder that indicates the ready list is empty

// // nodeT is channel message
// type nodeT struct {
// 	value int
// 	next  *nodeT
// }

// // Chan ...
// type Chan struct {
// 	waitg  sync.WaitGroup // used for sleeping. gotta get our zzz's
// 	queue  *nodeT         // items in the sender queue
// 	readys *nodeT         // items ready for receiving
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

// // Send sends a message of the receiver.
// func (ch *Chan) Send(value int) {
// 	n := &nodeT{value: value}
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
// 		ch.waitg.Done()
// 	}
// }

// // Recv receives the next message.
// func (ch *Chan) Recv() (value int) {
// 	// look in receiver list for items before checking the sender queue.
// 	for {
// 		readys := (*nodeT)(atomic.LoadPointer(
// 			(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
// 		))
// 		if readys == nil || readys == emptyN {
// 			break
// 		}
// 		if atomic.CompareAndSwapPointer(
// 			(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
// 			unsafe.Pointer(readys), unsafe.Pointer(readys.next)) {
// 			return readys.value
// 		}
// 		runtime.Gosched()
// 	}

// 	// let's load more messages from the sender queue.
// 	var queue *nodeT
// 	for {
// 		queue = ch.load()
// 		if queue == nil {
// 			// sender queue is empty. put the receiver to sleep
// 			ch.waitg.Add(1)
// 			if ch.cas(queue, sleepN) {
// 				ch.waitg.Wait()
// 			} else {
// 				ch.waitg.Done()
// 			}
// 		} else if ch.cas(queue, nil) {
// 			// empty the queue
// 			break
// 		}
// 		runtime.Gosched()
// 	}
// 	// reverse the order
// 	var prev *nodeT
// 	var current = queue
// 	var next *nodeT
// 	for current != nil {
// 		next = current.next
// 		current.next = prev
// 		prev = current
// 		current = next
// 	}
// 	value = prev.value
// 	if prev.next != nil {
// 		// we have ordered items that must be handled later
// 		for {
// 			readys := (*nodeT)(atomic.LoadPointer(
// 				(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
// 			))
// 			if readys != emptyN {
// 				queue.next = readys
// 			} else {
// 				queue.next = nil
// 			}
// 			if atomic.CompareAndSwapPointer(
// 				(*unsafe.Pointer)(unsafe.Pointer(&ch.readys)),
// 				unsafe.Pointer(readys), unsafe.Pointer(prev.next)) {
// 				break
// 			}
// 			runtime.Gosched()
// 		}
// 	}
// 	return value
// }
