package ipcheck

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type IPInfo struct {
	HTTP interface {
		Do(*http.Request) (*http.Response, error)
	}
	Token string
}

func (a IPInfo) Locate(ctx context.Context, ip string) (Location, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipinfo.io/"+ip+"/json", nil)
	if err != nil {
		return Location{}, err
	}
	req.Header.Set("Accept", "application/json")
	if a.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.Token)
	}
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return Location{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Location{}, errors.New("ipinfo: " + resp.Status)
	}
	var body struct {
		City, Region, Country, Loc, Org, Timezone string
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Location{}, errors.New("ipinfo: " + err.Error())
	}
	loc := Location{City: body.City, Region: body.Region, Country: body.Country, Org: body.Org, Timezone: body.Timezone}
	if lat, lon, ok := strings.Cut(body.Loc, ","); ok {
		loc.Lat, _ = strconv.ParseFloat(lat, 64)
		loc.Lon, _ = strconv.ParseFloat(lon, 64)
	}
	return loc, nil
}
