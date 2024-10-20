package reverse

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

func PrintReverseList(head *ListNode) {
	if head.Next == nil {
		fmt.Println(head.Val)
		return
	}

	PrintReverseList(head.Next)
	fmt.Println(head.Val)
}

func ReverseList(head *ListNode) *ListNode {
	if head == nil {
		return nil
	}

	slow, fast := head, head.Next
	for fast != nil {
		temp := fast.Next
		fast.Next = slow
		slow = fast
		fast = temp
	}

	return slow
}
