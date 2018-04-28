package fastlane

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/tidwall/lotsa"
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

func printResults(key string, N, P int, dur time.Duration) {
	key = (key + "" + strings.Repeat(" ", 60))[:17]
	s := fmt.Sprintf("%s %d ops in %dms (%d/sec)",
		key, N, int(dur.Seconds()*1000), int(float64(N)/dur.Seconds()))
	ss := "s"
	if P == 1 {
		ss = ""
	}
	fmt.Printf("%s %d producer%s\n", (s + strings.Repeat(" ", 100))[:60], P, ss)
}

func TestFastlaneChan(t *testing.T) {
	N := 1000000
	var start time.Time

	//P := 1
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkFastlaneChan(N, P, P == 1)
		printResults("fastlane-channel", N, P, time.Since(start))

		start = time.Now()
		benchmarkGoChan(N, 100, P, P == 1)
		printResults("go-channel(100)", N, P, time.Since(start))

		start = time.Now()
		benchmarkGoChan(N, 10, P, P == 1)
		printResults("go-channel(10)", N, P, time.Since(start))

		start = time.Now()
		benchmarkGoChan(N, 0, P, P == 1)
		printResults("go-channel(0)", N, P, time.Since(start))
	}

}

func benchmarkFastlaneChan(N int, P int, validate bool) {
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
	lotsa.Ops(N, P, func(i, _ int) {
		ch.Send(uint64(i))
	})
	wg.Wait()
}

func benchmarkGoChan(N, buffered int, producers int, validate bool) {
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
	lotsa.Ops(N, producers, func(i, _ int) {
		ch <- uint64(i)
	})
	wg.Wait()
}

func BenchmarkFastlaneChan(b *testing.B) {
	b.ReportAllocs()
	benchmarkFastlaneChan(b.N, 1, false)
}

func BenchmarkGoChan100(b *testing.B) {
	b.ReportAllocs()
	benchmarkGoChan(b.N, 100, 1, false)
}

func BenchmarkGoChan10(b *testing.B) {
	b.ReportAllocs()
	benchmarkGoChan(b.N, 10, 1, false)
}

func BenchmarkGoChanUnbuffered(b *testing.B) {
	b.ReportAllocs()
	benchmarkGoChan(b.N, 0, 1, false)
}
