package main

import (
	"fmt"
	"minisort/bubble"
	"minisort/insertion"
	"minisort/merge"
	"minisort/quick"
	"minisort/selection"
)

func main() {
	var arr = []int{3, 5, 2, 1, 0, -1}
	fmt.Println(bubble.BubbleSort(arr))
	fmt.Println(selection.SelectionSort(arr))
	fmt.Println(insertion.InsertionSort(arr))

	var arr1 = []int{3, 5, 2, 6, 0, 1, 9, 2}
	fmt.Println(merge.MergeSort(arr1))

	quick.QuickSort(arr1)
	fmt.Println(arr1)

	var arr2 = []int{3, 2, 1, 5, 6, 4}
	fmt.Println(quick.FindKthLargest(arr2, 2))
}
