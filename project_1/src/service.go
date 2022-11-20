// Package satservice is a satelitte web service written in Go and deployed on the Google Cloud Service.
// 	1) GET request with location (latitude, longitude) as query parameters: e.g. https://tvao-178408.appspot.com/images?lat=40.457375&lng=-80.009353
// 	2) GET request with address using the Google Geocoding API: https://tvao-178408.appspot.com/images?address=Rued Langgaards Vej 7
// 		Returns JSON array containing links (i.e. granule ids) to all satellite images for that location
// 		NB: Marshalling may be used to map between JSON and Go values and encoding between JSON and a stream in HTTP request
//	DEPLOYMENT
//  	Development: "dev_appserver.py .""
// 		Public: "gcloud app deploy app.yaml" (deploys app to Google Cloud App Engine)
// 		Browse: "gcloud app browse"
// 		NB: remember to add your google cloud project ID as an environment variable in /Users/<user>/.bash_profile --> "export PROJECT_ID_GO=<project-id>"
package satservice

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
)

// init is run before the application starts serving
func init() {
	http.HandleFunc("/", redirect)
	http.Handle("/images", appHandler(images))
}

// redirect ensures that client is redirected to correct route
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://tvao-178408.appspot.com/images", 301)
}

// Basic regular expressions for validating user input and column number for granules
const (
	Latitude  string = "^[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)$"
	Longitude string = "^[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)$"
)

// Define custom HTTP appHandler that includes error return value to reduce repetition in error handling
type appHandler func(http.ResponseWriter, *http.Request) *appError

// User friendly error representation with error, message and HTTP status code
type appError struct {
	Error   error
	Message string
	Code    int // Server (500 Internal Error) or Client (400 Bad Request Error)
}

// Implement ServeHTTP to comply with the http.Handler interface
// Go functional feature: fn is a first order function that invokes the underlying http request function (e.g. get)
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := fn(w, r); err != nil {
		http.Error(w, err.Message, err.Code)
	}
}

// Exercise 2 and 4: Returns JSON array with links to all satellite images (i.e. granule ids) based on a location
// Location is based on a latitude and longitude or address provided as query parameters
func images(w http.ResponseWriter, r *http.Request) *appError {
	if err := r.ParseForm(); err != nil {
		return &appError{err, "Cannot parse data", http.StatusInternalServerError}
	}

	address := r.Form.Get("address")
	lat, lng, err := convertAddressToCoords(address, r)

	if err != nil {
		lat, lng = r.Form.Get("lat"), r.Form.Get("lng")
	}

	validLat, validLng := regexp.MustCompile(Latitude).MatchString(lat), regexp.MustCompile(Longitude).MatchString(lng)

	if !validLat && !validLng {
		return &appError{errors.New("Invalid coordinates"), "Please provide a valid latitude and longitude", http.StatusBadRequest}
	}

	projectID := "tvao-178408" // Cloud project ID used by BigQuery API - TODO: replace with os.Getenv("GO_PROJECT_ID")
	links, err := getLinks(lat, lng, projectID, r)
	if err != nil {
		return &appError{err, "Unable to retrieve links", http.StatusInternalServerError}
	}

	if err := json.NewEncoder(w).Encode(links); err != nil {
		return &appError{err, "Unable to map JSON to response", http.StatusInternalServerError}
	}

	log.Printf("Success: granule links fetched from latitude '%s' and longitude '%s'", lat, lng)
	return nil // Success
}
