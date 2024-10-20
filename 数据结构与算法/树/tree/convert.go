package tree

import "fmt"

/*
	func findLeftMaxNode(root *TreeNode) *TreeNode {
		cur := root
		for cur != nil {
			cur = cur.Right
		}

		return cur
	}

	func findRightMinNode(root *TreeNode) *TreeNode {
		cur := root
		for cur != nil {
			cur = cur.Left
		}

		return cur
	}

	func Convert(root *TreeNode) *TreeNode {
		if root == nil {
			return nil
		}

		if root.Left != nil {
			Convert(root.Left)
			if lmax := findLeftMaxNode(root.Left); lmax != nil {
				root.Left = lmax
				lmax.Right = root
			}
		}

		if root.Right != nil {
			Convert(root.Right)
			if rmin := findRightMinNode(root.Right); rmin != nil {
				root.Right = rmin
				rmin.Left = root
			}
		}

		return root
	}
*/

var head, pre, tail *TreeNode = nil, nil, nil

func midOrderWithList(root *TreeNode) {
	if root == nil {
		return
	}

	midOrderWithList(root.Left)

	if head == nil {
		head = root
	}

	if pre == nil {
		pre = root
	} else {
		pre.Right = root
		root.Left = pre
		pre = root
	}

	tail = root

	midOrderWithList(root.Right)
}

func Convert(root *TreeNode) (*TreeNode, *TreeNode) {
	if root == nil {
		return nil, nil
	}

	midOrderWithList(root)

	return head, tail
}

func PrintRightNodes(root *TreeNode) {
	for cur := root; cur != nil; cur = cur.Right {
		fmt.Printf("%d ", cur.Val)
	}
	fmt.Println()
}

func PrintLeftNodes(root *TreeNode) {
	for cur := root; cur != nil; cur = cur.Left {
		fmt.Printf("%d ", cur.Val)
	}
	fmt.Println()
}
