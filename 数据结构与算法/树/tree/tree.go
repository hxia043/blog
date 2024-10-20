package tree

type BinarySearchTree struct {
	root *Node
}

type Node struct {
	data  int
	left  *Node
	right *Node
}

func (t *BinarySearchTree) Remove(data int) bool {
	if t.root == nil {
		return false
	}

	pre, cur := t.root, t.root
	for cur != nil {
		if cur.data == data {
			if cur.left == nil && cur.right == nil {
				if pre.left == cur {
					pre.left = nil
				} else {
					pre.right = nil
				}
			} else if cur.left == nil && cur.right != nil {
				if pre.left == cur {
					pre.left = cur.right
				}
				pre.right = cur.right
			} else if cur.left != nil && cur.right == nil {
				if pre.left == cur {
					pre.left = cur.left
				}
				pre.right = cur.left
			} else {
				minpp, minp := cur, cur.right
				for minp.left != nil {
					minpp = minp
					minp = minp.left
				}

				if minpp.left == minp {
					minpp.left = minp.right
				} else if minpp.right == minp {
					minpp.right = minp.right
				}

				cur.data = minp.data
				minp.right = nil
			}

			return true
		}

		pre = cur
		if cur.data > data {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	return false
}

func (t *BinarySearchTree) FindPreNode(data int) *Node {
	if t.root == nil || t.root.data == data {
		return nil
	}

	pre, cur := t.root, t.root
	for cur != nil {
		if cur.data == data {
			return pre
		}

		pre = cur
		if cur.data > data {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	return pre
}

func (t *BinarySearchTree) FindNextNode(data int) (*Node, *Node) {
	if t.root == nil {
		return nil, nil
	}

	cur := t.root
	for cur != nil {
		if cur.data == data {
			return cur.left, cur.right
		}

		if cur.data > data {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	return nil, nil
}

func (t *BinarySearchTree) Find(data int) bool {
	if t.root == nil {
		return false
	}

	cur := t.root
	for cur != nil {
		if cur.data == data {
			return true
		}

		if cur.data > data {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	return false
}

func (t *BinarySearchTree) Insert(data int) {
	if t.root == nil {
		t.root = &Node{data: data}
		return
	}

	cur := t.root
	for {
		if cur.data > data {
			if cur.left == nil {
				cur.left = &Node{data: data}
				break
			}
			cur = cur.left
		} else {
			if cur.right == nil {
				cur.right = &Node{data: data}
				break
			}
			cur = cur.right
		}
	}
}

func (t *BinarySearchTree) Paths(val int) [][]int {
	var totalPath = [][]int{}

	q := NewQueue2()
	q.Push([]int{t.root.data}, t.root)

	for q.Len() > 0 {
		paths := q.Pop().Val()

		sum := 0
		for _, path := range paths {
			sum += path
		}

		if sum == val {
			totalPath = append(totalPath, paths)
		}

		cur := q.Pop().Node()
		if cur.left != nil {
			left := append(paths, cur.left.data)
			q.Push(left, cur.left)
		}

		if cur.right != nil {
			right := append(paths, cur.right.data)
			q.Push(right, cur.right)
		}

		q.Remove()
	}

	return totalPath
}

func New() *BinarySearchTree {
	return &BinarySearchTree{
		root: nil,
	}
}
