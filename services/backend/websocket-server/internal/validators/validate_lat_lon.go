package validators

func LatLonWithinUnitedStates(lat float64, lon float64) (valid bool) {
	if lat > 49.3457868 || lat < 24.7433195 {
		return false
	}
	if lon > -66.9513812 || lon < -124.7844079 {
		return false
	}
	return true
}
