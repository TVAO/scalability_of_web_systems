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
package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof" // Profiling
	"regexp"
	"strings"
)

// init is run before the application starts serving
func init() {
	http.HandleFunc("/", redirect)
	http.Handle("/images", appHandler(images))
	http.Handle("/area", appHandler(area))
}

// use this for local testing with go tool
// func main() {
// 	appengine.Main()
// 	http.HandleFunc("/", redirect)
// 	http.Handle("/images", appHandler(images))
// 	http.Handle("/area", appHandler(area))
// }

// redirect ensures that client is redirected to correct route
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://tvao-178408.appspot.com/area", 301)
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

// Project 1 - Exercise 2 and 4: Returns JSON array with links to all satellite images (i.e. granule ids) based on a location
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

// Project 2 : Image data in geographic location
// Returns a JSON array with links to all satellite images within a marked area of interest specified with a pair of lat/lng coordinates.
// Area of interest is specified by a pair of latitude and longitude coordinates as query parameters.
func area(w http.ResponseWriter, r *http.Request) *appError {
	if err := r.ParseForm(); err != nil {
		return &appError{err, "Cannot parse data", http.StatusInternalServerError}
	}

	projectID := "tvao-178408"
	lat1, lng1, lat2, lng2 := r.Form.Get("lat1"), r.Form.Get("lng1"), r.Form.Get("lat2"), r.Form.Get("lng2")
	if !regexp.MustCompile(Latitude).MatchString(lat1) || !regexp.MustCompile(Latitude).MatchString(lat2) ||
		!regexp.MustCompile(Longitude).MatchString(lng1) || !regexp.MustCompile(Longitude).MatchString(lng2) {
		return &appError{errors.New("Invalid coordinates"), "Please provide a valid pair of latitude and longitude bands \n" +
			" Example: https://tvao-178408.appspot.com/area?lat1=55.698473&lng1=12.506052&lat2=55.616879&lng2=12.652524", http.StatusBadRequest}
	}

	links, err := getImageBaseURL(lat1, lng1, lat2, lng2, projectID, r)
	if err != nil {
		return &appError{err, "Unable to retrieve links", http.StatusInternalServerError}
	}

	// Create a set of worker jobs for each link
	numberOfJobs := len(links)
	jobs := make(chan string, numberOfJobs)
	results := make(chan Links, numberOfJobs)

	// Setup worker pool
	for i := 0; i <= numberOfJobs; i++ {
		go worker(r, jobs, results)
	}

	// Send jobs
	for _, imgLink := range links {
		jobs <- imgLink
	}
	close(jobs) // Close do indicate this is all work to be done

	// Collect worker results and write them to JSON result
	imageResult := Links{}
	for i := 0; i <= numberOfJobs; i++ {
		imageResult = append(imageResult, <-results...)
	}
	close(results)

	// Encode JSON result
	encodeErr := json.NewEncoder(w).Encode(imageResult)
	if encodeErr != nil {
		log.Fatal("Error")
	}

	return nil // Success
}

// Worker receives work on jobs channel and send images for each folder job to result
func worker(r *http.Request, jobs <-chan string, results chan<- Links) {
	folderImages := Links{}
	for imgLink := range jobs {
		linkAndGranule := strings.SplitAfter(imgLink, "gcp-public-data-sentinel-2")
		bucketName := linkAndGranule[0]
		imageObject := strings.Trim(linkAndGranule[1], "/")
		result, err := getImagesFromBucket(bucketName, imageObject, r)

		if err != nil {
			log.Fatalln("Error on worker")
		}
		folderImages = result
	}
	results <- folderImages
}
