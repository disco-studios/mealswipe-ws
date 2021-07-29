package foursquare

import "sort"

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

type By func(p1, p2 *Venue) bool

func (by By) Sort(venues []Venue) {
	vs := &venueSorter{
		venues: venues,
		by:     by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(vs)
}

type venueSorter struct {
	venues []Venue
	by     func(p1, p2 *Venue) bool
}

func (s *venueSorter) Len() int {
	return len(s.venues)
}

func (s *venueSorter) Swap(i, j int) {
	s.venues[i], s.venues[j] = s.venues[j], s.venues[i]
}

func (s *venueSorter) Less(i, j int) bool {
	return s.by(&s.venues[i], &s.venues[j])
}

func Distance(v1, v2 *Venue) bool {
	return v1.Location.Distance < v2.Location.Distance
}
