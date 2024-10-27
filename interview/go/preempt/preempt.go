package preempt

import "fmt"

func Process() {
	for {
		for i := 0; i < 1e6; i++ {
			empty()
		}
	}
}

func empty() {
	fmt.Println("empty")
}
