package tree

import (
	"fmt"
)

func preOrder(n *Node) {
	if n == nil {
		return
	}

	fmt.Printf("%d ", n.data)
	preOrder(n.left)
	preOrder(n.right)
}

func (t *BinarySearchTree) PreOrder() {
	preOrder(t.root)
	fmt.Println()
}

func midOrder(n *Node) {
	if n == nil {
		return
	}

	midOrder(n.left)
	fmt.Printf("%d ", n.data)
	midOrder(n.right)
}

func (t *BinarySearchTree) MidOrder() {
	midOrder(t.root)
	fmt.Println()
}

func postOrder(n *Node) {
	if n == nil {
		return
	}

	postOrder(n.left)
	postOrder(n.right)
	fmt.Printf("%d ", n.data)
}

func (t *BinarySearchTree) PostOrder() {
	postOrder(t.root)
	fmt.Println()
}

func (t *BinarySearchTree) Sequence() {
	q := NewQueue()

	cur := t.root
	q.Push(cur)

	for q.Len() > 0 {
		cur = q.Pop().Val()
		q.Remove()

		fmt.Printf("%d ", cur.data)
		if cur.left != nil {
			q.Push(cur.left)
		}
		if cur.right != nil {
			q.Push(cur.right)
		}
	}

	fmt.Println()
}

func midKthOrder(n *Node, k int, ticker *int, number *int) {
	if n == nil {
		return
	}

	midKthOrder(n.left, k, ticker, number)
	*ticker = *ticker + 1
	if *ticker == k {
		*number = n.data
		return
	}

	midKthOrder(n.right, k, ticker, number)
}

func (t *BinarySearchTree) FindKthNumber(k int) int {
	ticker, number := 0, 0
	midKthOrder(t.root, k, &ticker, &number)

	return number
}
