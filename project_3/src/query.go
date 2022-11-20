// package satservice is used to query satellite image data with the BigQuery and Storage REST API in Google Cloud Services
package satservice

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
)

// Position of granule id column in table
const (
	granuleIDColumn = 0
	baseURLColumn   = 1
	projectID       = "tvao-178408" // TODO: os.GetEnv()
)

// Links encapsulates the links (i.e. granule ids)  fetched from Google Cloud via BigQuery
type Links []string

// Retrieves links (i.e. granule ids) of all satellite images via a location based on a latitude and longitude
func getLinks(lat, lng string, r *http.Request) (Links, error) {
	granuleQuery := strings.TrimSpace(fmt.Sprintf(
		`SELECT granule_id
		 FROM %[1]sbigquery-public-data.cloud_storage_geo_index.sentinel_2_index%[1]s
		 WHERE %[2]s < north_lat
		 AND south_lat < %[2]s
		 AND %[3]s < east_lon
		 AND west_lon < %[3]s;`, "`", lat, lng))

	var links Links
	ctx := appengine.NewContext(r)
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return links, err
	}

	query := client.Query(granuleQuery)
	query.QueryConfig.UseStandardSQL = true
	rows, err := query.Read(ctx)

	for {
		var row []bigquery.Value
		err := rows.Next(&row) // No rows left
		if err == iterator.Done {
			return links, nil // Returns result
		}
		if err != nil {
			return nil, err
		}

		granuleID := row[granuleIDColumn].(string)
		links = append(links, granuleID)
	}
}

// Project 2 : Image data in geographic location
// Fetches all sentinel-2 image folders that contain image data within the specified area of interest, using the Big Query Api
func getImageBaseURL(lat1, lng1, lat2, lng2 string, r *http.Request) (Links, error) {
	imageURLQuery := strings.TrimSpace(fmt.Sprintf(
		`SELECT base_url, granule_id 
		FROM %[1]sbigquery-public-data.cloud_storage_geo_index.sentinel_2_index%[1]s
		WHERE %[2]s < north_lat
		AND south_lat < %[4]s
		AND %[3]s < east_lon
		AND west_lon < %[5]s;`, "`", lat1, lng1, lat2, lng2)) // Argument 2, 3, 4, 5
	links := Links{}
	client, err := bigquery.NewClient(r.Context(), projectID)
	if err != nil {
		return nil, err
	}

	query := client.Query(imageURLQuery)
	query.QueryConfig.UseStandardSQL = true
	rows, err := query.Read(r.Context())
	if err != nil {
		return nil, err
	}

	row := []bigquery.Value{}
	imageBaseURL, granuleID, fullImageURL := "", "", ""
	for {
		err := rows.Next(&row) // No rows left
		if err == iterator.Done {
			return links, nil // Returns result
		}
		if err != nil {
			return nil, err
		}
		imageBaseURL = strings.Replace(row[0].(string), "gs://", "", 1) // Removes trailing gs:// from bucket name
		granuleID = row[1].(string)
		fullImageURL = imageBaseURL + "/GRANULE/" + granuleID + "/IMG_DATA/"
		links = append(links, fullImageURL)
	}
}

// Project 3 : Fetch all links to granules containing a subfolder of images that match specified area of interest, using Big query API
// This version works in parallel by using goroutines and channels
// TODO: refactor getImageBaseUrl to support setting concurrency level for fetching links in parallel
func getImageCount(client *bigquery.Client, r *http.Request, channel chan int, errors chan error, lat1, lng1, lat2, lng2 string) {
	count := 0
	imageURLQuery := strings.TrimSpace(fmt.Sprintf(
		`SELECT COUNT(granule_id)  
		FROM %[1]sbigquery-public-data.cloud_storage_geo_index.sentinel_2_index%[1]s
		WHERE %[2]s < north_lat
		AND south_lat < %[4]s
		AND %[3]s < east_lon
		AND west_lon < %[5]s;`, "`", lat1, lng1, lat2, lng2))

	query := client.Query(imageURLQuery)
	query.QueryConfig.UseStandardSQL = true
	rows, err := query.Read(r.Context())
	if err != nil {
		errors <- err
	}

	row := []bigquery.Value{}
	for {
		err := rows.Next(&row) // No rows left
		if err == iterator.Done {
			channel <- count // Write image count to channel instead of returning
			break
		}
		if err != nil {
			errors <- err
		}
		imgCount := int(row[0].(int64))
		count += imgCount
	}
}

// Project 2 : Image data in geographic location
// Fetches a complete list of image ids from a specified image folder in the sentinel-2 folder, using the Cloud Bucket Storage API
func getImagesFromBucket(client *storage.Client, bucketName, objectName string, r *http.Request) (Links, error) {
	query := storage.Query{Prefix: objectName, Versions: false}
	links := Links{}
	fullImageURL := bytes.Buffer{}

	it := client.Bucket(bucketName).Objects(r.Context(), &query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		fullImageURL.WriteString(bucketName + "/" + attrs.Name)
		links = append(links, fullImageURL.String())
		fullImageURL.Reset()
	}
	return links, nil
}

// Project 3 : Helper function (https://www.dotnetperls.com/duplicates-go)
// Remove potentially duplicated granules in region cover to avoid counting images twice
// func removeDuplicatesUnordered(elements []string) []string {
// 	encountered := map[string]bool{}

// 	// Create a map of all unique elements.
// 	for v := range elements {
// 		encountered[elements[v]] = true
// 	}

// 	// Place all keys from the map into a slice.
// 	result := []string{}
// 	for key := range encountered {
// 		result = append(result, key)
// 	}
// 	return result
// }
