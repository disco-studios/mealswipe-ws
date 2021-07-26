package foursquare

type VenueLocation struct {
	Distance int
	Lat      float64
	Lng      float64
	Address  string
}

type VenuePhoto struct {
	Prefix string
	Suffix string
	Width  int
	Height int
}

type VenuePhotosGroup struct {
	Items []VenuePhoto
}

type VenuePhotosBody struct {
	Groups []VenuePhotosGroup
}

type Venue struct {
	Id       string
	Name     string
	Photos   VenuePhotosBody
	Location VenueLocation
}

type LocationRequestResponseBody struct {
	Venues []Venue
}

type LocationRequestResponse struct {
	Response LocationRequestResponseBody
}

type VenueRequestResponseBody struct {
	Venue Venue
}

type VenueRequestResponse struct {
	Response VenueRequestResponseBody
}
