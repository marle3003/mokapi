package faker

func Set[T any](array []T, i int, val T) ([]T, func() []T) {
	var restore func() []T
	result := array
	if i < len(array) {
		old := array[i]
		array[i] = val
		restore = func() []T {
			array[i] = old
			return array
		}
	} else {
		result = make([]T, i)
		copy(result, array)
		result = append(result, val)
		restore = func() []T {
			return array
		}
	}
	return result, restore
}

func splice[T any](array []T, start int, deleteCount int, items []T) ([]T, func() []T) {
	if start < 0 {
		return array, nil
	}

	end := start + deleteCount
	if end > len(array) {
		end = len(array)
	}

	var toAdd []T
	for _, item := range items {
		toAdd = append(toAdd, item)
	}

	removed := array[start:end]

	result := make([]T, 0, len(array)+len(toAdd)-deleteCount)
	result = append(result, array[:start]...)
	result = append(result, toAdd...)
	result = append(result, array[end:]...)

	restore := func() []T {
		restore := make([]T, start)
		copy(restore, result[:start])

		restore = append(restore, removed...)

		added := len(toAdd)
		restore = append(restore, result[start+added:]...)
		return restore
	}

	return result, restore
}
