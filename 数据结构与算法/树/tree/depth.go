package tree

func Depth(n *Node2) int {
	if n == nil {
		return 0
	}

	leftDepth := Depth(n.Left)
	rightDepth := Depth(n.Right)

	if leftDepth > rightDepth {
		return leftDepth + 1
	}

	return rightDepth + 1
}
