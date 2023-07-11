package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

func sha(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return fmt.Sprintf("%x", sum)
}

func calc(start int, end int, netID string) (int, bool) {
	for i := start; i < end; i++ {
		payload := fmt.Sprintf("%s %d\n", netID, i)
		shasum := sha(payload)
		if strings.HasPrefix(shasum, "00000") {
			return i, true
		}
	}

	return 0, false
}

func parallel(limit int, netID string) (int, bool) {
	c := make(chan int, 1)
	done := make(chan struct{}, 1)

	bucketSize := limit / 4
	wg := &sync.WaitGroup{}
	for i := 0; i < limit; i += bucketSize {
		wg.Add(1)
		go func(i int) {
			target, ok := calc(i, i+bucketSize, netID)
			if ok {
				c <- target
			}
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case target := <-c:
		return target, true
	case <-done:
		return 0, false
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("usage: <net-id> <limit>\n")
		return
	}
	netID := os.Args[1]
	limit, _ := strconv.Atoi(os.Args[2])
	target, ok := parallel(limit, netID)
	if ok {
		fmt.Printf("target:[%d]\n", target)
	} else {
		fmt.Printf("target not found in range [0:%d]\n", limit)
	}
}
