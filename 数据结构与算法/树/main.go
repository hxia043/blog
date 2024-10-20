package main

import (
	"fmt"
	"tree/tree"
)

func main() {
	root := tree.New()
	root.Insert(13)
	root.Insert(10)
	root.Insert(16)
	root.Insert(9)
	root.Insert(11)
	root.Insert(14)

	if root.Find(11) {
		fmt.Println("find 11 in the tree")
	}

	if root.Find(12) {
		fmt.Println("find 12 in the tree")
	}

	preNode := root.FindPreNode(11)
	fmt.Println(preNode)
	preNode = root.FindPreNode(12)
	fmt.Println(preNode)

	nextLeftNode, nextRightNode := root.FindNextNode(10)
	fmt.Println(nextLeftNode, nextRightNode)
	nextLeftNode, nextRightNode = root.FindNextNode(12)
	fmt.Println(nextLeftNode, nextRightNode)

	root.PreOrder()
	root.MidOrder()

	fmt.Println(root.FindKthNumber(5))

	root.PostOrder()

	//root.Remove(9)
	//root.MidOrder()

	//root.Remove(16)
	//root.MidOrder()

	//root.Remove(10)
	//root.MidOrder()

	root.MidOrder()
	root.Sequence()

	// 示例数据：前序和中序遍历
	preorder := []int{1, 2, 4, 7, 3, 5, 6, 8}
	inorder := []int{4, 7, 2, 1, 5, 3, 8, 6}

	// 构建二叉树
	r := tree.BuildTree(preorder, inorder)

	// 输出中序遍历，验证重建结果
	fmt.Print("Inorder traversal of the rebuilt tree: ")
	tree.InorderTraversal(r)
	fmt.Println()

	a := &tree.Node2{Data: "a"}
	b := &tree.Node2{Data: "b"}
	c := &tree.Node2{Data: "c"}
	d := &tree.Node2{Data: "d"}
	e := &tree.Node2{Data: "e"}
	f := &tree.Node2{Data: "f"}
	g := &tree.Node2{Data: "g"}
	h := &tree.Node2{Data: "h"}
	i := &tree.Node2{Data: "i"}

	a.Left, a.Right = b, c
	b.Left, b.Right = d, e
	c.Left, c.Right = f, g
	e.Left, e.Right = h, i

	b.Pre, c.Pre = a, a
	d.Pre, e.Pre = b, b
	h.Pre, i.Pre = e, e
	f.Pre, g.Pre = c, c

	fmt.Println(tree.FindNext(d))
	fmt.Println(tree.FindNext(b))
	fmt.Println(tree.FindNext(e))
	fmt.Println(tree.FindNext(i))
	fmt.Println(tree.FindNext(a))
	fmt.Println(tree.FindNext(g))

	fmt.Println(tree.Depth(a))

	root2 := tree.New()
	root2.Insert(10)
	root2.Insert(5)
	root2.Insert(12)
	root2.Insert(4)
	root2.Insert(7)

	fmt.Println(root2.Paths(22))

	//root3 := &tree.TreeNode0{}
	//fmt.Println(tree.IsValidBST(root3))

	root4 := &tree.TreeNode0{}
	h1 := &tree.TreeNode0{Val: -1}
	root4.Left = h1

	fmt.Println(tree.IsValidBST(root4))

	n4 := &tree.TreeNode{Val: 4}
	n2 := &tree.TreeNode{Val: 2}
	n7 := &tree.TreeNode{Val: 7}
	n1 := &tree.TreeNode{Val: 1}
	n3 := &tree.TreeNode{Val: 3}
	n6 := &tree.TreeNode{Val: 6}
	n9 := &tree.TreeNode{Val: 9}
	n4.Left, n4.Right = n2, n7
	n2.Left, n2.Right = n1, n3
	n7.Left, n7.Right = n6, n9

	fmt.Println(tree.InvertTree(n4))

	tree.InorderTraversal(n4)
	fmt.Println()
	tree.Image(n4)
	tree.InorderTraversal(n4)
	fmt.Println()

	h10 := &tree.TreeNode{Val: 10}
	h6 := &tree.TreeNode{Val: 6}
	h14 := &tree.TreeNode{Val: 14}
	h4 := &tree.TreeNode{Val: 4}
	h8 := &tree.TreeNode{Val: 8}
	h12 := &tree.TreeNode{Val: 12}
	h16 := &tree.TreeNode{Val: 16}
	h10.Left, h10.Right = h6, h14
	h6.Left, h6.Right = h4, h8
	h14.Left, h14.Right = h12, h16

	head, tail := tree.Convert(h10)
	tree.PrintRightNodes(head)
	tree.PrintLeftNodes(tail)
}
