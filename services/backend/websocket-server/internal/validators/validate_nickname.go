package validators

func IsNicknameValid(nickname string) (valid bool, err error) {
	if len(nickname) == 0 || len(nickname) > 16 {
		return false, nil
	}
	return true, nil
}
