package validators

import (
	"testing"
)

// Valid nickname:
//  - 3 <= length <= 16
//  - a-zA-Z
//  - ^(?!\s)([a-zA-Z]*)(?<!\s)$

func TestIsNicknameValid(t *testing.T) {
	nickname := ""

	t.Run("Lead space invalid", func(t *testing.T) {
		nickname = " Cam the Man"
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("Trail space invalid", func(t *testing.T) {
		nickname = "Cam the Man "
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("NonAlphanumeric Invalid", func(t *testing.T) {
		nickname = "C@m the Man "
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("More than 2 spaces in a row invalid", func(t *testing.T) {
		nickname = "Cam  the Man "
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("Using other whitespace invalid", func(t *testing.T) {
		nickname = "Cam\tthe Man "
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("Too long invalid", func(t *testing.T) {
		nickname = "Cam the Mannnnnnnnnnnn"
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("Empty invalid", func(t *testing.T) {
		nickname = ""
		if valid, err := IsNicknameValid(nickname); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("One char valid", func(t *testing.T) {
		nickname = "a"
		if valid, err := IsNicknameValid(nickname); !valid || err != nil {
			t.FailNow()
		}
	})

	t.Run("Max length valid", func(t *testing.T) {
		nickname = "aaaaaaaaaaaaaaaa"
		if valid, err := IsNicknameValid(nickname); !valid || err != nil {
			t.FailNow()
		}
	})

	t.Run("Valid nickname", func(t *testing.T) {
		nickname = "Cam the Man"
		if valid, err := IsNicknameValid(nickname); !valid || err != nil {
			t.FailNow()
		}
	})

}
