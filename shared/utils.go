package shared

func ValidateCLI(inputs ...interface{}) bool {
	for _, input := range inputs {
		if input == nil {
			return false
		}

		if str, ok := input.(string); ok {
			return len(str) > 0
		}
	}

	return true
}
