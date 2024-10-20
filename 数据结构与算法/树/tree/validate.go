package tree

type TreeNode0 struct {
	Val   int
	Left  *TreeNode0
	Right *TreeNode0
}

var midorders = make([]int, 0)

func midOrder2(root *TreeNode0) {
	if root == nil {
		return
	}

	midOrder2(root.Left)
	midorders = append(midorders, root.Val)
	midOrder2(root.Right)
}

func IsValidBST(root *TreeNode0) bool {
	if root == nil {
		return true
	}

	midOrder2(root)

	for i := 0; i+1 < len(midorders); i++ {
		if midorders[i] >= midorders[i+1] {
			return false
		}
	}

	return true
}
