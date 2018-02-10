package fastlane

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestFastlaneChan(t *testing.T) {
	N := 1000000
	var start time.Time
	var dur time.Duration

	start = time.Now()
	benchmarkFastlaneChan(N, true)
	dur = time.Since(start)
	fmt.Printf("fastlane-channel: %d ops in %s (%d/sec)\n", N, dur, int(float64(N)/dur.Seconds()))

	start = time.Now()
	benchmarkGoChan(N, true)
	dur = time.Since(start)
	fmt.Printf("go-channel:       %d ops in %s (%d/sec)\n", N, dur, int(float64(N)/dur.Seconds()))
}

func benchmarkFastlaneChan(N int, validate bool) {
	var ch Chan
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v := ch.Recv()
			if validate {
				if v.(int) != i {
					panic("out of order")
				}
			}
		}
		wg.Done()
	}()
	for i := 0; i < N; i++ {
		ch.Send(i)
	}
	wg.Wait()
}
func benchmarkGoChan(N int, validate bool) {
	ch := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < N; i++ {
			v := <-ch
			if validate {
				if v != i {
					panic("out of order")
				}
			}
		}
		wg.Done()
	}()
	for i := 0; i < N; i++ {
		ch <- i
	}
	wg.Wait()
}

func BenchmarkFastlaneChan(b *testing.B) {
	benchmarkFastlaneChan(b.N, false)
}

func BenchmarkGoChan(b *testing.B) {
	benchmarkGoChan(b.N, false)
}
