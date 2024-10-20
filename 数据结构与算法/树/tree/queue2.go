package tree

type Queue2 struct {
	head *node2
	tail *node2
}

type node2 struct {
	data []int
	cur  *Node
	next *node2
}

func (n *node2) Val() []int {
	return n.data
}

func (n *node2) Node() *Node {
	return n.cur
}

func (q *Queue2) Pop() *node2 {
	if q.head == nil {
		return nil
	}

	return q.head
}

func (q *Queue2) Len() int {
	len := 0
	for cur := q.head; cur != nil; cur = cur.next {
		len++
	}

	return len
}

func (q *Queue2) Remove() {
	if q.head == q.tail {
		q.head, q.tail = nil, nil
		return
	}

	temp := q.head.next
	q.head.next = nil
	q.head = temp
}

func (q *Queue2) Push(data []int, cur *Node) {
	n := &node2{data: data, cur: cur}
	if q.head == nil {
		q.head, q.tail = n, n
		return
	}

	q.tail.next = n
	q.tail = n
}

func NewQueue2() *Queue2 {
	return &Queue2{}
}
