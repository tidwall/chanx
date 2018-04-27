package chan2

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

var sleepN = &nodeT{} // placeholder that indicates the receiver is sleeping
var emptyN = &nodeT{} // placeholder that indicates the ready list is empty

// nodeT is channel message
type nodeT struct {
	value int
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
func (ch *Chan) Send(value int) {
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
func (ch *Chan) Recv() (value int) {
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
	value = prev.value
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
	return value
}
