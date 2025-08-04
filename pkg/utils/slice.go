package utils

// RemoveDuplicatesInOrder removes duplicate elements from a slice while preserving the original order.
// It works for any slice of comparable type T.
// Example: RemoveDuplicatesInOrder([]int{1,2,2,3}) returns []int{1,2,3}
func RemoveDuplicatesInOrder[T comparable](arr []T) []T {
	encountered := make(map[T]bool)
	result := make([]T, 0, len(arr)) // Preallocate capacity to avoid multiple allocations

	for _, v := range arr {
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return result
}
