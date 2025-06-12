package mokapi

var Delete = &struct{}{}

func patch(target, patch any) any {
	if patch == nil {
		return target
	}
	if target == nil {
		return patch
	}

	mapTarget, isTargetMap := target.(map[string]any)
	mapPatch, isPatchMap := patch.(map[string]any)

	if isTargetMap && isPatchMap {
		return patchMap(mapTarget, mapPatch)
	} else {
		arrTarget, isTargetArray := target.([]any)
		arrPatch, isPatchArray := patch.([]any)
		if isTargetArray && isPatchArray {
			return patchArray(arrTarget, arrPatch)
		} else {
			return patch
		}
	}
}

func patchMap(t, p map[string]any) map[string]any {
	result := make(map[string]any)

	// copy original value
	for k, v := range t {
		result[k] = v
	}

	for k, v := range p {
		if v == Delete {
			delete(result, k)
			continue
		}

		if vt, ok := t[k]; ok {
			result[k] = patch(vt, v)
		} else {
			result[k] = v
		}
	}
	return result
}

func patchArray(t, p []any) []any {
	result := make([]any, 0, len(p))

	// copy original value
	for _, v := range t {
		result = append(result, v)
	}

	for i, v := range p {
		if v == Delete {
			result = append(result[:i], result[i+1:]...)
			continue
		}
		if i < len(result) {
			result[i] = patch(result[i], v)
		} else {
			result = append(result, v)
		}
	}

	return result
}
