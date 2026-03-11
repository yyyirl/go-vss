/**
 * @Author:         yi
 * @Description:    set_test
 * @Version:        1.0.0
 * @Date:           2025/2/13 19:36
 */
package set

import (
	"sync"
	"testing"
)

// go test -bench=. -benchmem
func Benchmark_add_test(b *testing.B) {
	var (
		set = New[int](0)
		wg  sync.WaitGroup
	)
	for i := 0; i < 100000000; i++ {
		wg.Add(1)

		go func(i int) {
			set.Add(i)
			defer wg.Done()
		}(i)
	}

	wg.Wait()
}

func Benchmark_remove_test(b *testing.B) {
	var (
		set = New[int](0)
		wg  sync.WaitGroup
	)
	for i := 0; i < 10000000; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			set.Remove(i)
		}(i)
	}

	wg.Wait()
}
