package channel

import (
	"fmt"
	"sync"
	"time"
)

type task struct {
	t  string
	id int
}

func worker(id int, taskChan <-chan task, wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range taskChan {
		if t.t == "event" {
			fmt.Printf("worker %d receive the task %s from %d\n", id, t.t, t.id)
		}

		if t.t == "log" {
			fmt.Printf("worker %d receive the task %s from %d\n", id, t.t, t.id)
		}
	}
}

func productor(id int, taskChan chan<- task, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(10 * time.Second)
	for i := 0; i < 10; i++ {
		taskChan <- task{id: id, t: "event"}
		taskChan <- task{id: id, t: "log"}
	}
}

func CSP() {
	var taskChan = make(chan task, 10)
	var productorWg sync.WaitGroup

	for i := 0; i < 2; i++ {
		productorWg.Add(1)
		go productor(i, taskChan, &productorWg)
	}

	go func() {
		productorWg.Wait()
		close(taskChan)
	}()

	var workerWg sync.WaitGroup
	for i := 0; i < 3; i++ {
		workerWg.Add(1)
		go worker(i, taskChan, &workerWg)
	}

	workerWg.Wait()
}
