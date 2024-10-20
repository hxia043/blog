package insertion

func InsertionSort(arr []int) []int {
	n := len(arr)
	for i := 0; i < n; i++ {
		if arr[i] < arr[0] {
			temp := arr[i]
			for j := i; j >= 1; j-- {
				arr[j] = arr[j-1]
			}
			arr[0] = temp
		}
	}

	return arr
}
