package common

import "testing"

func TestIsRadiusValid(t *testing.T) {
	t.Run("Less than 500m invalid", func(t *testing.T) {
		if valid, err := IsRadiusValid(0); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("Negative invalid", func(t *testing.T) {
		if valid, err := IsRadiusValid(-1000); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("Greater than 20000m invalid", func(t *testing.T) {
		if valid, err := IsRadiusValid(20001); valid && err == nil {
			t.FailNow()
		}
	})

	t.Run("500m valid", func(t *testing.T) {
		if valid, err := IsRadiusValid(500); err != nil || !valid {
			t.FailNow()
		}
	})

	t.Run("20000m valid", func(t *testing.T) {
		if valid, err := IsRadiusValid(20000); err != nil || !valid {
			t.FailNow()
		}
	})

	t.Run("1000m valid", func(t *testing.T) {
		if valid, err := IsRadiusValid(1000); err != nil || !valid {
			t.FailNow()
		}
	})
}
