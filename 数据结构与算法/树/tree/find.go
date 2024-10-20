package tree

type Node2 struct {
	Data  string
	Pre   *Node2
	Left  *Node2
	Right *Node2
}

func FindNext(n *Node2) *Node2 {
	if n.Right == nil {
		for n.Pre != nil {
			if n.Pre.Left == n {
				return n.Pre
			}

			n = n.Pre
		}

		return nil
	}

	min := n.Right
	for min.Left != nil {
		min = min.Left
	}
	return min
}
