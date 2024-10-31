package channel

import (
	"fmt"
	"sync"
)

func Channel3() {
	var wg sync.WaitGroup
	var numChan = make(chan int)
	var charChan = make(chan rune)

	wg.Add(2)
	go func() {
		defer wg.Done()
		var c = 'A'
		for n := range numChan {
			fmt.Print(n)
			charChan <- c
			c++
		}

		close(charChan)
	}()

	go func() {
		defer wg.Done()
		var n = 1
		for c := range charChan {
			fmt.Print(c)
			if c == 'Z' {
				fmt.Println()
				close(numChan)
				continue
			}
			numChan <- n
			n++
		}
	}()

	wg.Wait()
}
