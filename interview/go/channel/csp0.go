package channel

import (
	"fmt"
	"sync"
)

func worker0(id int, taskChan <-chan task, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case t := <-taskChan:
			if t.t == "event" {
				fmt.Printf("worker %d receive the task %s from %d\n", id, t.t, t.id)
			}

			if t.t == "log" {
				fmt.Printf("worker %d receive the task %s from %d\n", id, t.t, t.id)
			}
		case <-stopChan:
			return
		}
	}
}

func productor0(id int, taskChan chan<- task, stopChan <-chan struct{}, toStop chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		if i == 8 {
			if _, ok := <-stopChan; ok {
				return
			}

			toStop <- id
			return
		}

		select {
		case <-stopChan:
			return
		case taskChan <- task{id: id, t: "event"}:
		case taskChan <- task{id: id, t: "log"}:
		}
	}
}

func CSP0() {
	var taskChan = make(chan task, 10)
	var wg sync.WaitGroup
	var toStop = make(chan int)
	var stopChan = make(chan struct{})

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go productor0(i, taskChan, stopChan, toStop, &wg)
	}

	go func() {
		for range toStop {
			close(stopChan)
			return
		}
	}()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go worker0(i, taskChan, stopChan, &wg)
	}

	wg.Wait()
}
