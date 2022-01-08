package shared

func ValidateArgs(inputs ...interface{}) (bool, []string) {
	var errs []string

	for _, input := range inputs {
		if input == nil {
			errs = append(errs, "input cannot be nil")
		}

		if str, ok := input.(string); ok {
			if len(str) <= 0 {
				errs = append(errs, "string input cannot be empty")
			}
		}
	}

	if len(errs) > 0 {
		return false, errs
	}

	return true, nil
}
