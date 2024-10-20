package main

import (
	"find/arr"
	"find/find"
	"find/rotate"
	"find/sqrt"
	"fmt"
)

func main() {
	var arr0 = []int{4, 5, 6, 7, 0, 1, 2}
	fmt.Println(rotate.FindMin(arr0))

	var arr1 = []int{1, 2, 3, 3, 3, 3, 4, 5}
	fmt.Println(find.FindCount(arr1, 3))

	var arr2 = []int{3, 3, 3, 3, 3, 3, 3, 3}
	fmt.Println(find.FindCount(arr2, 3))

	fmt.Println(sqrt.Sqrt(8))

	var arr3 = []int{0, 1, 2, 3, 5, 6, 7}
	fmt.Println(arr.GetMissingNumber(arr3))

	var arr4 = []int{-3, -1, 1, 3, 5}
	fmt.Println(arr.GetEqualNumber(arr4))
}
