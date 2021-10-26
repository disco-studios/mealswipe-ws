package common

import "mealswipe.app/mealswipe/internal/codes"

func IsCodeValid(code string) (valid bool) {
	if len(code) != 6 {
		return false
	}

	// Make sure each char in the code is valid
	for _, char := range code {
		found := false
		for _, charsetChar := range codes.SESSION_CODE_CHARSET {
			if string(char) == charsetChar {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
