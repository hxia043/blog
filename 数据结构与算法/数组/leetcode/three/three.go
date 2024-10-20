package three

import "sort"

func ThreeSum(nums []int) [][]int {
	n := len(nums)
	right := n - 1

	sort.Ints(nums)
	threes := [][]int{}
	for i := 0; i < n-2; i++ {
		target := -1 * nums[i]
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}

		for left := i + 1; left < n && left != right; left++ {
			if left > i+1 && nums[left] == nums[left-1] {
				continue
			}

			if nums[left]+nums[right] > target {
				left, right = left-1, right-1
			} else if nums[left]+nums[right] == target {
				threes = append(threes, []int{nums[i], nums[left], nums[right]})
			}
		}
	}

	return threes
}
