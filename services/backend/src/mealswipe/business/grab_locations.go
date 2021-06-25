package business

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type FS_Venue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FS_Venues_Stepper struct {
	Venues []*FS_Venue `json:"venues"`
}

type FS_Venues_Result struct {
	Venues FS_Venues_Stepper `json:"response"`
}

func GrabLocations(lat float32, lng float32) (venues []*FS_Venue, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.foursquare.com/v2/venues/search?client_id=UIEPSPWBZLULKZJQGT3KNRBX40O4GHBKA1SZ404HCMTUYCSN&client_secret=3QD0PJNSFOJTWWLZCGO3ERHCTQEVA4L11LSEFFDLAOKFSDVR&v=20210620&ll=%f,%f&intent=browse&radius=1000&limit=50&categoryId=4bf58dd8d48988d16e941735", lat, lng))
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res FS_Venues_Result
	err = json.Unmarshal(bodyBytes, &res)
	if err != nil {
		return nil, err
	}

	return res.Venues.Venues, nil
}
