package duplicate

func Duplicate(arr []int) int {
	N := len(arr) - 1
	m := 0
	for i := 0; i < N; i++ {
		for arr[i] != i {
			m = arr[i]
			if arr[m] == m {
				return m
			}

			arr[m], arr[i] = arr[i], arr[m]
		}
	}

	return 0
}
