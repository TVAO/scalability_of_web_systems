// Package satservice converter provides functionality to convert a human-like addres to a coordinate via the Google Geolocation API
// Geocoding is the process of converting an address into its geographic lat/lng coordinates
package satservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// JSON result returned by Geolocation API
type geoResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64
				Lng float64
			}
		}
	}
}

// Converts a human-like address to coordinates (latitude and longitude) via the Google Geolocation API
// A Google Maps Geocoding API request has the form: https://maps.googleapis.com/maps/api/geocode/json?address=<address>,
// where output is json and the required parameter is an address
func convertAddressToCoords(address string, r *http.Request) (string, string, error) {

	if address == "" {
		return "", "", errors.New("Invalid address input")
	}

	safeAddress := url.QueryEscape(address) // Escapes string so it is safe to place inside URL query

	// Geocoding API
	fullURL := fmt.Sprintf("http://maps.googleapis.com/maps/api/geocode/json?address=%s", safeAddress)

	// App engine context to interact with external service via http client
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	response, err := client.Get(fullURL)

	if err != nil {
		return "", "", err
	}

	defer response.Body.Close()

	// Generate latitude and longitude from address using Google Geocoding API
	// Use json.Decode or json.Encode for reading or writing streams of JSON data
	var res geoResponse
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return "", "", err
	}

	lat := strconv.FormatFloat(res.Results[0].Geometry.Location.Lat, 'f', 6, 64)
	lng := strconv.FormatFloat(res.Results[0].Geometry.Location.Lng, 'f', 6, 64)
	log.Printf("Success: converted address '%s' into lat = '%s' and lng = '%s' \n", address, lat, lng)

	return lat, lng, nil // Success
}
