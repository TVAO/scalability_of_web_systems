/*

 */

package satservice

import (
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine"
)

// Position of granule id column in table
const baseGranuleColumn = 0

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

		granuleID := row[baseGranuleColumn].(string)
		links = append(links, granuleID)
	}
}
