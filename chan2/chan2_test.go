package chan2

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tidwall/lotsa"
)

const producers = 1
const N = 1000000

func TestABCGo(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan int, 128)

	// senders
	all := make([]int, N)
	var count int32

	start := time.Now()
	go func() {
		// senders
		lotsa.Ops(N, producers, func(i, _ int) {
			ch <- i
		})
		ch <- -1
	}()

	lastV := -1
	for {
		v := <-ch
		if v == -1 {
			break
		}
		if v < lastV {
			panic("out of order")
		}
		n := int(atomic.AddInt32(&count, 1))
		all[n-1] = v
		if n == N {
			dur := time.Since(start)
			sort.Ints(all)
			for i := 0; i < N; i++ {
				if all[i] != i {
					panic("bad data")
				}
			}
			fmt.Printf("%d ops in %.0fms (%.0f ops/sec) (%d ns/op)\tGo\n", N, dur.Seconds(), float64(N)/dur.Seconds(), dur/time.Duration(N))
		}
	}
}

func TestFastlane(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	var ch Chan

	// senders
	all := make([]int, N)
	var count int32

	start := time.Now()
	go func() {
		// senders
		lotsa.Ops(N, producers, func(i, _ int) {
			ch.Send(i)
		})
		ch.Send(-1)
	}()

	lastV := -1
	for {
		v := ch.Recv()
		if v == -1 {
			return
		}
		if v <= lastV {
			panic("out of order")
		}
		n := int(atomic.AddInt32(&count, 1))
		all[n-1] = v
		if n == N {
			dur := time.Since(start)
			// dur := time.Since(start)
			sort.Ints(all)
			for i := 0; i < N; i++ {
				if all[i] != i {
					panic("bad data")
				}
			}
			fmt.Printf("%d ops in %.0fms (%.0f ops/sec) (%d ns/op)\tFastlane\n", N, dur.Seconds(), float64(N)/dur.Seconds(), dur/time.Duration(N))
		}
	}

}
