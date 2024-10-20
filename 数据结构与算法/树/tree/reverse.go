package tree

func InvertTree(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}

	temp := root.Left
	root.Left = root.Right
	root.Right = temp

	InvertTree(root.Left)
	InvertTree(root.Right)

	return root
}
