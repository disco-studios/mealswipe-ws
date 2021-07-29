package validators

func IsRadiusValid(radius int32) (valid bool, err error) {
	// Greater than or equal to .5km
	// Less than or equal to 20km
	return 500 <= radius && radius <= 20000, nil
}
