package question

import (
	"fmt"
	"sync"
	"testing"
)

func TestMap(t *testing.T) {
	cm := NewConcurrentMapByChannel()
	defer cm.Close()

	var wg sync.WaitGroup
	// 并发写
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cm.Set(fmt.Sprintf("key-%d", n), n)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			value := cm.Get(key)
			fmt.Printf("Key: %s, Value: %v\n", key, value)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Map size: %d\n", cm.Len())
}
