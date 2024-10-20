package tree

import "fmt"

// 定义二叉树节点
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 构建二叉树的主函数
func BuildTree(preorder []int, inorder []int) *TreeNode {
	if len(preorder) == 0 || len(inorder) == 0 {
		return nil
	}

	// 前序遍历的第一个元素是根节点
	rootVal := preorder[0]
	root := &TreeNode{Val: rootVal}

	// 找到根节点在中序遍历中的位置
	var rootIndex int
	for i, val := range inorder {
		if val == rootVal {
			rootIndex = i
			break
		}
	}

	// 构建左子树和右子树
	root.Left = BuildTree(preorder[1:1+rootIndex], inorder[:rootIndex])
	root.Right = BuildTree(preorder[1+rootIndex:], inorder[rootIndex+1:])

	return root
}

// 中序遍历打印二叉树
func InorderTraversal(root *TreeNode) {
	if root != nil {
		InorderTraversal(root.Left)
		fmt.Print(root.Val, " ")
		InorderTraversal(root.Right)
	}
}
