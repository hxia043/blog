package merge

func merge(left []int, right []int) []int {
	var sorted = []int{}

	p1, p2 := 0, 0
	for {
		if p1 == len(left) {
			sorted = append(sorted, right[p2:]...)
			return sorted
		}

		if p2 == len(right) {
			sorted = append(sorted, left[p1:]...)
			return sorted
		}

		if left[p1] > right[p2] {
			sorted = append(sorted, right[p2])
			p2++
		} else {
			sorted = append(sorted, left[p1])
			p1++
		}
	}
}

func sort(left []int, right []int) []int {
	if len(left) == 1 || len(right) == 1 {
		return merge(left, right)
	}

	lk := len(left) / 2
	rk := len(right) / 2

	lsort := sort(left[:lk], left[lk:])
	rsort := sort(right[:rk], right[rk:])

	return merge(lsort, rsort)
}

func MergeSort(arr []int) []int {
	k := len(arr) / 2
	return sort(arr[:k], arr[k:])
}
