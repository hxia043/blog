package main

import (
	"book/duplicate"
	"fmt"
)

func main() {
	var a = []int{2, 3, 1, 0, 2, 5, 3}
	fmt.Println(duplicate.Duplicate(a))
}
