package main

import (
	"bytes"
	"fmt"
	"log"
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
)

// Links encapsulates the links (i.e. granule ids)  fetched from Google Cloud via BigQuery
type Links []string

// Retrieves links (i.e. granule ids) of all satellite images via a location based on a latitude and longitude
func getLinks(lat, lng, proj string, r *http.Request) (Links, error) {
	granuleQuery := strings.TrimSpace(fmt.Sprintf(
		`SELECT granule_id
		 FROM %[1]sbigquery-public-data.cloud_storage_geo_index.sentinel_2_index%[1]s
		 WHERE %[2]s < north_lat
		 AND south_lat < %[2]s
		 AND %[3]s < east_lon
		 AND west_lon < %[3]s;`, "`", lat, lng))

	var links Links
	ctx := appengine.NewContext(r)
	client, err := bigquery.NewClient(ctx, proj)
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
			return links, err
		}

		granuleID := row[granuleIDColumn].(string)
		links = append(links, granuleID)
	}
}

// Project 2 : Image data in geographic location
// Fetches all sentinel-2 image folders that contain image data within the specified area of interest, using the Big Query Api
func getImageBaseURL(lat1, lng1, lat2, lng2, proj string, r *http.Request) (Links, error) {
	imageURLQuery := strings.TrimSpace(fmt.Sprintf(
		`SELECT base_url, granule_id 
		FROM %[1]sbigquery-public-data.cloud_storage_geo_index.sentinel_2_index%[1]s
		WHERE %[2]s < north_lat
		AND south_lat < %[4]s
		AND %[3]s < east_lon
		AND west_lon < %[5]s;`, "`", lat1, lng1, lat2, lng2)) // Argument 2, 3, 4, 5

	links := Links{}
	ctx := appengine.NewContext(r)
	client, err := bigquery.NewClient(ctx, proj)
	if err != nil {
		return nil, err
	}

	query := client.Query(imageURLQuery)
	query.QueryConfig.UseStandardSQL = true
	rows, err := query.Read(ctx)

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

// Project 2 : Image data in geographic location
// Fetches a complete list of image ids from a specified image folder in the sentinel-2 folder, using the Cloud Bucket Storage API
func getImagesFromBucket(bucketName, objectName string, r *http.Request) (Links, error) {
	ctx := appengine.NewContext(r)
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return nil, err
	}

	query := storage.Query{Prefix: objectName, Versions: false}
	links := Links{}
	fullImageURL := bytes.Buffer{}

	it := client.Bucket(bucketName).Objects(ctx, &query)
	for {
		attrs, err := it.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		fullImageURL.WriteString(bucketName + "/" + attrs.Name)
		links = append(links, fullImageURL.String())
		fullImageURL.Reset()
	}
	return links, nil
}
