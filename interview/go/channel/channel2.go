package channel

import (
	"fmt"
	"sync"
)

func Channel2() {
	var switchCh = make(chan bool)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i <= 26; i++ {
			<-switchCh
			fmt.Print(i)
			switchCh <- true
		}
	}()

	go func() {
		defer wg.Done()
		switchCh <- true
		for i := 'A'; i <= 'Z'; i++ {
			<-switchCh
			fmt.Print(string(i))
			if i == 'Z' {
				fmt.Println()
				close(switchCh)
				return
			}
			switchCh <- true
		}
	}()

	wg.Wait()
}
