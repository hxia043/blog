package main

import (
	"linklist/reverse"
)

func main() {
	h5 := &reverse.ListNode{
		Val:  5,
		Next: nil,
	}
	h4 := &reverse.ListNode{
		Val:  4,
		Next: h5,
	}
	h3 := &reverse.ListNode{
		Val:  3,
		Next: h4,
	}
	h2 := &reverse.ListNode{
		Val:  2,
		Next: h3,
	}
	h1 := &reverse.ListNode{
		Val:  1,
		Next: h2,
	}

	//fmt.Println(reverse.ReverseList(h1))
	reverse.PrintReverseList(h1)
}
