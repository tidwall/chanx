package chanx

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tidwall/lotsa"
)

var xbench bool

func init() {
	xbench = os.Getenv("CHANXBENCH") == "1"
	if !xbench {
		fmt.Printf("Use CHANXBENCH=1 for benchmarks\n")
	}
}

func TestOrder(t *testing.T) {
	// testing that order is preserved
	type msgT struct{ i, thread int }
	ch := Make(0)
	N := 1000000
	T := 100
	go func() {
		lotsa.Ops(N, 100, func(i, thread int) {
			if !ch.Send(&msgT{i, thread}) {
				panic("bad news")
			}
		})
		ch.Close()
	}()
	// create unique buckets per thread and store each message
	// sequentially in their respective bucket.
	m := make(map[int][]int)
	for {
		v, ok := ch.Recv()
		if !ok {
			break
		}
		msg := v.(*msgT)
		m[msg.thread] = append(m[msg.thread], msg.i)
	}
	// check that each bucket contains ordered data check for duplicates
	all := make(map[int]bool)
	for thread := 0; thread < T; thread++ {
		b, ok := m[thread]
		// println(thread, m[thread])
		// continue
		if !ok {
			t.Fatal("missing bucket")
		}
		if len(b) != N/T {
			t.Fatal("invalid bucket size")
		}
		h := -1
		for i := 0; i < len(b); i++ {
			if b[i] <= h {
				t.Fatal("out of order")
			}
			h = b[i]
			if all[h] {
				t.Fatal("duplicate value")
			}
			all[h] = true
		}
	}
}

func fixLeft(s string, n int) string {
	return (s + strings.Repeat(" ", n))[:n]
}
func fixRight(s string, n int) string {
	return (strings.Repeat(" ", n) + s)[len(s):]
}

func printResults(key string, N, P int, dur time.Duration) {
	s := fixLeft(key, 13) + " "
	s += fixLeft(fmt.Sprintf("%d ops in %dms", N, int(dur.Seconds()*1000)), 22) + " "
	s += fixRight(fmt.Sprintf("%d/sec", int(float64(N)/dur.Seconds())), 12) + " "
	s += fixRight(fmt.Sprintf("%dns/op", int(dur/time.Duration(N))), 10) + " "
	s += fixRight(fmt.Sprintf("%s %4d producer", (s + strings.Repeat(" ", 100))[:60], P), 14)
	fmt.Printf("%s\n", strings.TrimSpace(s))
}

func TestChanxUnbuffered(t *testing.T) {
	N := 1000000
	for P := 1; P < 1000; P *= 10 {
		start := time.Now()
		benchmarkChanx(N, 0, P, false)
		if xbench {
			printResults("chanx(0)", N, P, time.Since(start))
		}
	}
}

func TestChanx10Unbuffered(t *testing.T) {
	N := 1000000
	for P := 1; P < 1000; P *= 10 {
		start := time.Now()
		benchmarkChanx(N, 10, P, false)
		if xbench {
			printResults("chanx(10)", N, P, time.Since(start))
		}
	}
}

func TestChanx100Unbuffered(t *testing.T) {
	N := 1000000
	for P := 1; P < 1000; P *= 10 {
		start := time.Now()
		benchmarkChanx(N, 100, P, false)
		if xbench {
			printResults("chanx(100)", N, P, time.Since(start))
		}
	}
}

func TestGoChanUnbuffered(t *testing.T) {
	if !xbench {
		return
	}
	N := 1000000
	var start time.Time
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkGoChan(N, 0, P, false)
		printResults("go-chan(0)", N, P, time.Since(start))
	}
}

func TestGoChan10(t *testing.T) {
	if !xbench {
		return
	}
	N := 1000000
	var start time.Time
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkGoChan(N, 10, P, false)
		printResults("go-chan(10)", N, P, time.Since(start))
	}
}

func TestGoChan100(t *testing.T) {
	if !xbench {
		return
	}
	N := 1000000
	var start time.Time
	for P := 1; P < 1000; P *= 10 {
		start = time.Now()
		benchmarkGoChan(N, 100, P, false)
		printResults("go-chan(100)", N, P, time.Since(start))
	}
}

func benchmarkChanx(N int, buffered int, P int, validate bool) {
	ch := Make(buffered)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v, _ := ch.Recv()
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

func Benchmark100ProducerChanx100(b *testing.B) {
	b.ReportAllocs()
	benchmarkChanx(b.N, 100, 100, false)
}

func Benchmark100ProducerChanx10(b *testing.B) {
	b.ReportAllocs()
	benchmarkChanx(b.N, 10, 100, false)
}

func Benchmark100ProducerChanxUnbuffered(b *testing.B) {
	b.ReportAllocs()
	benchmarkChanx(b.N, 0, 100, false)
}

func Benchmark100ProducerGoChan100(b *testing.B) {
	b.ReportAllocs()
	benchmarkGoChan(b.N, 100, 100, false)
}

func Benchmark100ProducerGoChan10(b *testing.B) {
	b.ReportAllocs()
	benchmarkGoChan(b.N, 10, 100, false)
}

func Benchmark100ProducerGoChanUnbuffered(b *testing.B) {
	b.ReportAllocs()
	benchmarkGoChan(b.N, 0, 100, false)
}
