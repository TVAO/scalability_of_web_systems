// package satservice allows you to query satellite images for meaningful areas (e.g. country)
// This is done using Geofabrik for polygons of countried, specified as a Planar straight-line graph (PSLG)
// The S2 geometry library is used to manipulate these geometrip shapes like Earth
package satservice

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"cloud.google.com/go/bigquery"

	"google.golang.org/appengine/urlfetch"

	"github.com/golang/geo/s2"
)

// Match floating point numbers with exponents using a regular expression
// Match optional - or +, 0 or more digits, optional ".", 0 or more digits, exponent, - or +, 0 or more digits
// Example: Longitude "8.552884E+00" and Latitude "5.491803E+01"
const (
	floatExponentPattern = "[-+]?[0-9]*\\.[0-9]+([eE][-+]?[0-9]+)?"
	bucketGranuleSize    = 13 // TODO: Fetch size dynamically via API call to storage client
)

// normalizeCoords is a helper function returns new slice containing result
// of "normalizing" (i.e. removing the exponent) in parsed coordinates
// Credits: https://gobyexample.com/collection-functions
func normalizeCoords(vs []string, f func(string, int) (float64, error)) ([]float64, error) {
	vsm := make([]float64, len(vs))
	for i, v := range vs {
		f, err := f(v, -1)
		if err != nil {
			return nil, err
		}
		vsm[i] = f //strconv.FormatFloat(f, 'f', -1, 64)
	}
	return vsm, nil
}

// Fetch and parse PSLG data from Geofabrik, based on a country specified by the user
func parse(r *http.Request, country, continent string) ([]float64, error) {
	client := urlfetch.Client(r.Context())
	request := ""

	if len(continent) > 0 {
		request = fmt.Sprintf("http://download.geofabrik.de/%s/%s.poly", continent, country)
	} else {
		request = fmt.Sprintf("http://download.geofabrik.de/%s.poly", country)
	}
	resp, err := client.Get(request)
	// Retry if error
	if err != nil {
		err := retry(DefaultRetry().MaxRetries, DefaultRetry().Duration*time.Second, func() (err error) {
			resp, err = client.Get(request)
			return
		})
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	regex := regexp.MustCompile(floatExponentPattern)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	data := regex.FindAllString(string(bytes), -1)
	countryCoords, err := normalizeCoords(data, strconv.ParseFloat)
	if err != nil {
		return nil, err
	}
	return countryCoords, nil
}

// Construct region cover from polygon, based on country coords
// Region of country is approximated as unions of cells (CellUnion)
// MaxLevel determines the granularity of cells covering regions, where 30 = 0,48 cm^2
// MaxCells determines how many cells are used to cover the given region
func regionCover(coords []float64, maxLevel, maxCells int) s2.CellUnion {
	// Parse coordinates into points
	points := []s2.Point{}
	for len(coords) > 0 {
		lat, lng := coords[0], coords[1]
		p := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lng))
		points = append(points, p)
		coords = coords[2:] // Rest coords
	}
	// Construct loop representing spherical polygon and polygon from loop
	l1 := s2.LoopFromPoints(points)
	loops := []*s2.Loop{l1}
	poly := s2.PolygonFromLoops(loops)
	// Construct region cover
	rc := &s2.RegionCoverer{MaxLevel: maxLevel, MaxCells: maxCells}
	cover := rc.Covering(poly)
	return cover
}

// Count satellite images associated to a country based on its polygon representation
// Use region cover data in combination with "query.go" to query relevant images with the Storage bucket API
func imagesByRegion(cover s2.CellUnion, r *http.Request) (int, error) {
	numberOfJobs := len(cover)
	results := make(chan int, numberOfJobs)
	errChan := make(chan error)
	imageCount := 0

	client, err := bigquery.NewClient(r.Context(), projectID)
	if err != nil {
		return 0, err
	}
	// Fetch image base links in parallel
	for i := 0; i < len(cover); i++ {
		c := s2.CellFromCellID(cover[i])
		go getImageCount(client, r, results, errChan,
			c.RectBound().Lo().Lat.String(),
			c.RectBound().Lo().Lng.String(),
			c.RectBound().Hi().Lat.String(),
			c.RectBound().Hi().Lng.String())
	}
	// Await concurrent results on channel
	for range cover {
		select {
		case err := <-errChan:
			return 0, err
		case count := <-results:
			imageCount += count
		}
	}
	close(results)
	log.Printf("Granules in region cover: %v", imageCount)
	return imageCount * bucketGranuleSize, nil
}

// Returns count of images within bounding box of country (for testing)
// func imagesByBox(rect s2.Rect, r *http.Request) (Links, error) {
// 	links, err := getImageBaseURL(rect.Lo().Lat.String(),
// 		rect.Lo().Lng.String(),
// 		rect.Hi().Lat.String(),
// 		rect.Hi().Lng.String(), r)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return links, nil
// }
