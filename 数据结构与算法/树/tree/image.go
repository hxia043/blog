package tree

import "fmt"

func Image(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}

	root.Left, root.Right = root.Right, root.Left

	Image(root.Left)
	Image(root.Right)

	return root
}

func midOrder0(n *TreeNode) {
	if n == nil {
		return
	}

	midOrder0(n.Left)
	fmt.Printf("%d ", n.Val)
	midOrder0(n.Right)
}

func (t *TreeNode) MidOrder() {
	midOrder0(t)
	fmt.Println()
}
