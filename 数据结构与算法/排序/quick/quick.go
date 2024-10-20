package quick

// partition 函数用于将数组划分为两部分：左边部分大于基准，右边部分小于基准
func partition(nums []int, left, right int) int {
	pivot := nums[right] // 基准元素
	i := left
	for j := left; j < right; j++ {
		if nums[j] > pivot { // 寻找第 K 大元素，使用降序排列
			nums[i], nums[j] = nums[j], nums[i]
			i++
		}
	}
	nums[i], nums[right] = nums[right], nums[i]
	return i
}

// quickselect 函数通过递归在数组中找到第 K 大元素
func quickselect(nums []int, left, right, k int) int {
	if left == right { // 当数组中只有一个元素时
		return nums[left]
	}

	pivotIndex := partition(nums, left, right)

	if pivotIndex == k { // 基准元素恰好是第 K 大元素
		return nums[pivotIndex]
	} else if pivotIndex > k { // 第 K 大元素在左边部分
		return quickselect(nums, left, pivotIndex-1, k)
	} else { // 第 K 大元素在右边部分
		return quickselect(nums, pivotIndex+1, right, k)
	}
}

// findKthLargest 函数用于找到数组中的第 K 大元素
func FindKthLargest(nums []int, k int) int {
	// 在数组长度为 n 的情况下，第 k 大元素是第 n - k 个索引
	return quickselect(nums, 0, len(nums)-1, k-1)
}

func sort(arr []int) int {
	p, q := 0, len(arr)-1
	v := arr[p]
	qWalk := true
	for {
		if p == q {
			arr[p] = v
			break
		}

		if qWalk {
			if arr[q] < v {
				arr[p] = arr[q]
				p++
				qWalk = false
			} else {
				q--
			}
		} else {
			if arr[p] > v {
				arr[q] = arr[p]
				q--
				qWalk = true
			} else {
				p++
			}
		}
	}

	return p
}

func QuickSort(arr []int) {
	if len(arr) == 0 || len(arr) == 1 {
		return
	}

	p := sort(arr)

	QuickSort(arr[:p])
	QuickSort(arr[p+1:])
}
