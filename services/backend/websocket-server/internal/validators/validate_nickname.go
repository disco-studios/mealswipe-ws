package validators

import "regexp"

func IsNicknameValid(nickname string) (valid bool, err error) {
	if len(nickname) == 0 || len(nickname) > 16 {
		return false, nil
	}
	// - Does not start or end with a space
	// - Only contains a-zA-Z and space
	// - Can only have one space in a row
	return regexp.MatchString(`^([a-zA-Z]+ ?)*[a-zA-Z]$`, nickname)
}
