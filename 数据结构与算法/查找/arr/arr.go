package arr

func GetEqualNumber(arr []int) int {
	l, r := 0, len(arr)-1
	for l <= r {
		mid := l + (r-l)/2

		if arr[mid] > mid {
			r = mid - 1
		} else if arr[mid] < mid {
			l = mid + 1
		} else {
			return mid
		}
	}

	return -1
}

func GetMissingNumber(arr []int) int {
	l, r := 0, len(arr)-1
	ans := -1
	for l <= r {
		mid := l + (r-l)/2
		if arr[mid] == mid {
			l = mid + 1
		} else {
			ans = mid
			r = mid - 1
		}
	}

	return ans
}
