package feature

var (
	data = map[string]struct{}{}
)

func Enable(features []string) {
	for _, feature := range features {
		data[feature] = struct{}{}
	}
}

func IsEnabled(name string) bool {
	if _, ok := data[name]; ok {
		return true
	}
	return false
}

func Reset() {
	data = map[string]struct{}{}
}
