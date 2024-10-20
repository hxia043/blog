package find

func FindCount(arr []int, n int) int {
	low, high := 0, len(arr)-1
	// find the index of number n
	k := 0
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] > n {
			high = mid - 1
		} else if arr[mid] < n {
			low = mid + 1
		} else {
			k = mid
			break
		}
	}

	// find the index of the first n
	first := 0
	low, high = 0, k
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] > n {
			high = mid - 1
		} else if arr[mid] < n {
			low = mid + 1
		} else {
			if mid-1 == 0 {
				if arr[mid-1] != n {
					first = mid
				} else {
					first = mid - 1
				}
				break
			} else {
				if arr[mid-1] != n {
					first = mid
					break
				} else {
					first = mid - 1
				}
			}
		}
	}

	// find the index of the last n
	last := 0
	low, high = k, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] > n {
			high = mid - 1
		} else if arr[mid] < n {
			low = mid + 1
		} else {
			if mid+1 == len(arr)-1 {
				if arr[mid+1] != n {
					last = mid
				} else {
					last = mid + 1
				}
				break
			} else {
				if arr[mid+1] != n {
					last = mid
					break
				} else {
					low = mid + 1
				}
			}
		}
	}

	return last - first + 1
}
