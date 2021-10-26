package common

import (
	"testing"
)

func TestIsCodeValid(t *testing.T) {
	t.Run("Valid code", func(t *testing.T) {
		if valid := IsCodeValid("BDXFGH"); valid == false {
			t.FailNow()
		}
	})

	t.Run("Vowel fails", func(t *testing.T) {
		if valid := IsCodeValid("EDXFGH"); valid == true {
			t.FailNow()
		}
	})

	t.Run("Short fails", func(t *testing.T) {
		if valid := IsCodeValid("BDFGH"); valid == true {
			t.FailNow()
		}
	})

	t.Run("Long fails", func(t *testing.T) {
		if valid := IsCodeValid("BDFFFGH"); valid == true {
			t.FailNow()
		}
	})

	t.Run("Empty fails", func(t *testing.T) {
		if valid := IsCodeValid(""); valid == true {
			t.FailNow()
		}
	})
}
