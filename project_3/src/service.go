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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof" // Profiling
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"google.golang.org/appengine"
)

// RequestRetrySession represents a user session where requests may be retried to improve resiliency
type RequestRetrySession struct {
	MaxRetries int
	Duration   time.Duration
	//Backoff
	//Session
}

// NewRetry creates a new retry session based on a given max attempt count and duration for each attempt
func NewRetry(retries int, duration time.Duration) RequestRetrySession {
	retrySession := RequestRetrySession{}
	retrySession.MaxRetries = retries
	retrySession.Duration = duration
	return retrySession
}

// DefaultRetry returns parameters used by default to retry requests
func DefaultRetry() RequestRetrySession {
	return RequestRetrySession{MaxRetries: 5, Duration: 10}
}

// init is run before the application starts serving
func init() {
	http.HandleFunc("/", redirect)
	http.Handle("/images", appHandler(images))
	http.Handle("/area", appHandler(area))
	http.Handle("/geo", appHandler(geo))
}

// redirect ensures that client is redirected to correct route
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://tvao-178408.appspot.com/geo", 301)
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
	ctx := appengine.NewContext(r)
	ctxWithDeadline, cancel := context.WithTimeout(ctx, 5*time.Minute)
	if err := fn(w, r.WithContext(ctxWithDeadline)); err != nil {
		http.Error(w, err.Message, err.Code)
	}
	defer cancel() // Cancel ctx as soon as request returns
	defer r.Body.Close()
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

	links, err := getLinks(lat, lng, r)
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

	lat1, lng1, lat2, lng2 := r.Form.Get("lat1"), r.Form.Get("lng1"), r.Form.Get("lat2"), r.Form.Get("lng2")
	if !regexp.MustCompile(Latitude).MatchString(lat1) || !regexp.MustCompile(Latitude).MatchString(lat2) ||
		!regexp.MustCompile(Longitude).MatchString(lng1) || !regexp.MustCompile(Longitude).MatchString(lng2) {
		return &appError{errors.New("Invalid coordinates"), "Please provide a valid pair of latitude and longitude bands \n" +
			" Example: https://tvao-178408.appspot.com/area?lat1=55.698473&lng1=12.506052&lat2=55.616879&lng2=12.652524", http.StatusBadRequest}
	}

	links, err := getImageBaseURL(lat1, lng1, lat2, lng2, r)
	if err != nil {
		return &appError{err, "Unable to retrieve granulelinks", http.StatusInternalServerError}
	}

	imageResult := pool(links, r)
	if err := imageResult.Error; err != nil {
		return &appError{err, "Could not fetch pictures from granules", http.StatusInternalServerError}
	}
	// Encode JSON result
	encodeErr := json.NewEncoder(w).Encode(len(imageResult.Links))
	if encodeErr != nil {
		return &appError{err, "Unable to encode JSON", http.StatusInternalServerError}
	}
	return nil // Success
}

// Project 3 : Fetch and parse PSLG data of country user inputs from Geofabrik
// Returns count of images associated with bounding box of country
func geo(w http.ResponseWriter, r *http.Request) *appError {
	if err := r.ParseForm(); err != nil || !(len(r.Form.Get("country")) > 0) {
		return &appError{err, "Could not parse specified country location.", http.StatusBadRequest}
	}

	country := r.Form.Get("country")
	continent := r.Form.Get("continent")
	coords, err := parse(r, country, continent)
	if err != nil {
		return &appError{err, "Could not fetch PSLG data", http.StatusInternalServerError}
	}

	cover := regionCover(coords, 15, 100)
	imageCount, err := imagesByRegion(cover, r)
	if err != nil {
		return &appError{err, "Could not get granules", http.StatusInternalServerError}
	}

	encodeErr := json.NewEncoder(w).Encode(imageCount)
	if encodeErr != nil {
		return &appError{encodeErr, "Unable to find region cover", http.StatusInternalServerError}
	}
	return nil
}

// Result represents links and wraps errors that may occur
type Result struct {
	Links []string
	Error error
}

// Worker pool used to fetch images from subfolders in Google Cloud Bucket concurrently using goroutines
func pool(links Links, r *http.Request) Result {
	// Create a set of worker jobs for each link
	numberOfJobs := len(links)
	jobs := make(chan string)
	results := make(chan Result)
	imageResult := Result{}

	// Clients should be reused instead of created as needed. The methods of Client are safe for concurrent use by multiple goroutines.
	client, err := storage.NewClient(r.Context())
	if err != nil {
		imageResult.Error = err
		return imageResult // Error propagated
	}

	// Start goroutine workers
	for i := 0; i <= numberOfJobs; i++ {
		go worker(client, r, jobs, results)
	}

	// Send jobs
	for _, imgLink := range links {
		jobs <- imgLink
	}
	close(jobs) // Close do indicate this is all work to be done

	// Collect worker results and write them to JSON result
	for i := 0; i <= numberOfJobs; i++ {
		result := <-results
		imageResult.Links = append(imageResult.Links, result.Links...)
	}
	close(results)
	return imageResult
}

// Worker receives work on jobs channel and send images for each folder job to result
func worker(client *storage.Client, r *http.Request, jobs <-chan string, results chan<- Result) {
	folderImages := Result{}
	for imgLink := range jobs {
		linkAndGranule := strings.SplitAfter(imgLink, "gcp-public-data-sentinel-2")
		bucketName := linkAndGranule[0]
		imageObject := strings.Trim(linkAndGranule[1], "/")
		//bucketHandle := client.Bucket(bucketName)
		result, err := getImagesFromBucket(client, bucketName, imageObject, r)

		// Retry for better resilience
		if err != nil {
			err := retry(DefaultRetry().MaxRetries, DefaultRetry().Duration*time.Second, func() (err error) {
				result, err = getImagesFromBucket(client, bucketName, imageObject, r)
				return
			})
			if err != nil {
				folderImages.Error = err
			}
		}
		folderImages.Links = result
	}
	results <- folderImages
}

// Google Client API may fail in which we want to enforce a retry mechanism to improve the resiliency
// Credits: https://blog.abourget.net/en/2016/01/04/my-favorite-golang-retry-function/
// http://sethammons.com/post/pester/
func retry(attempts int, sleep time.Duration, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}
		/// Add randomness to prevent Thundering Herd: https://upgear.io/blog/simple-golang-retry-function/
		jitter := time.Duration(rand.Int63n(int64(sleep)))
		sleep = sleep + jitter/2
		time.Sleep(sleep)
		//log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
