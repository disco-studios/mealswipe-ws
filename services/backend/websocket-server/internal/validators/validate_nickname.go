package validators

func IsNicknameValid(nickname string) (valid bool, err error) {
	if len(nickname) == 0 || len(nickname) > 30 {
		return false, nil
	}
	return true, nil
}
