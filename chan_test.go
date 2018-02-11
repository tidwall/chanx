package fastlane

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type MyType struct {
	hello int
}

type ChanMyType struct{ base ChanPointer }

func (ch *ChanMyType) Send(value *MyType) {
	ch.base.Send(unsafe.Pointer(value))
}
func (ch *ChanMyType) Recv() *MyType {
	return (*MyType)(ch.base.Recv())
}

func TestFastlaneChan(t *testing.T) {
	N := 1000000
	var start time.Time
	var dur time.Duration

	start = time.Now()
	benchmarkFastlaneChan(N, true)
	dur = time.Since(start)
	fmt.Printf("fastlane-channel: %d ops in %s (%d/sec)\n", N, dur, int(float64(N)/dur.Seconds()))

	start = time.Now()
	benchmarkGoChan(N, 100, true)
	dur = time.Since(start)
	fmt.Printf("go-channel(100):  %d ops in %s (%d/sec)\n", N, dur, int(float64(N)/dur.Seconds()))

	start = time.Now()
	benchmarkGoChan(N, 10, true)
	dur = time.Since(start)
	fmt.Printf("go-channel(10):   %d ops in %s (%d/sec)\n", N, dur, int(float64(N)/dur.Seconds()))

	start = time.Now()
	benchmarkGoChan(N, 0, true)
	dur = time.Since(start)
	fmt.Printf("go-channel(0):    %d ops in %s (%d/sec)\n", N, dur, int(float64(N)/dur.Seconds()))

}

func benchmarkFastlaneChan(N int, validate bool) {
	var ch ChanUint64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v := ch.Recv()
			if validate {
				if v != uint64(i) {
					panic("out of order")
				}
			}
		}
		wg.Done()
	}()
	for i := 0; i < N; i++ {
		ch.Send(uint64(i))
	}
	wg.Wait()
}

func benchmarkGoChan(N, buffered int, validate bool) {
	ch := make(chan uint64, buffered)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v := <-ch
			if validate {
				if v != uint64(i) {
					panic("out of order")
				}
			}
		}
		wg.Done()
	}()
	for i := 0; i < N; i++ {
		ch <- uint64(i)
	}
	wg.Wait()
}

func BenchmarkFastlaneChan(b *testing.B) {
	benchmarkFastlaneChan(b.N, false)
}

func BenchmarkGoChan100(b *testing.B) {
	benchmarkGoChan(b.N, 100, false)
}

func BenchmarkGoChan10(b *testing.B) {
	benchmarkGoChan(b.N, 10, false)
}

func BenchmarkGoChanUnbuffered(b *testing.B) {
	benchmarkGoChan(b.N, 0, false)
}
