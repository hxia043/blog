package tree

type Queue struct {
	head *node
	tail *node
}

type node struct {
	data *Node
	next *node
}

func (n *node) Val() *Node {
	return n.data
}

func (q *Queue) Pop() *node {
	if q.head == nil {
		return nil
	}

	return q.head
}

func (q *Queue) Len() int {
	len := 0
	for cur := q.head; cur != nil; cur = cur.next {
		len++
	}

	return len
}

func (q *Queue) Remove() {
	if q.head == q.tail {
		q.head, q.tail = nil, nil
		return
	}

	temp := q.head.next
	q.head.next = nil
	q.head = temp
}

func (q *Queue) Push(N *Node) {
	n := &node{data: N}
	if q.head == nil {
		q.head, q.tail = n, n
		return
	}

	q.tail.next = n
	q.tail = n
}

func NewQueue() *Queue {
	return &Queue{}
}
