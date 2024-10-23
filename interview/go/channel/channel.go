package channel

import (
	"fmt"
	"sync"
)

func worker1(msgChan chan msg, wg *sync.WaitGroup) {
	defer wg.Done()

	var m = msg{id: 0, number: 0}
	msgChan <- m

	for m1 := range msgChan {
		fmt.Println(m1.id, m1.number)
		if m1.number == 100 {
			close(msgChan)
			break
		}

		m.number = m1.number + 1
		msgChan <- m
	}
}

func worker2(msgChan chan msg, wg *sync.WaitGroup) {
	defer wg.Done()

	var m = msg{id: 1, number: 0}
	for m0 := range msgChan {
		fmt.Println(m0.id, m0.number)
		if m0.number == 100 {
			close(msgChan)
			break
		}

		m.number = m0.number + 1
		msgChan <- m
	}
}

type msg struct {
	id     int
	number int
}

func Print() {
	var msgChan = make(chan msg)

	var wg sync.WaitGroup

	wg.Add(2)
	go worker1(msgChan, &wg)
	go worker2(msgChan, &wg)

	wg.Wait()
}
