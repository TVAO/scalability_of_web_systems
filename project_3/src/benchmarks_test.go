// Package satservice : this contains a sample of benchmark tests, used to evaluate the memory consumption and cpu usage of the image queries
package satservice

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"google.golang.org/appengine/aetest"
)

// Benchmark the image query that returns a range of granules related to a specified latitude and longitude
func BenchmarkImages(b *testing.B) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		b.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	// Create a request to pass to handler with query parameters
	req, err := inst.NewRequest("GET", "/images", nil)
	req.Form = url.Values{"lat": {"55.660797"}, "lng": {"12.5896"}}
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	//handler := http.Handler(appHandler(images))

	// Note that we repeat the operation we want to benchmark (in this case service.images) b.N times.
	// This value will be changed by go test until the resulting times are statistically significant.
	for i := 0; i < b.N; i++ {
		//handler.ServeHTTP(rr, req)
		images(rr, req)
	}
}

// Benchmark spatial query that returns all image links within geographical area of interest
func BenchmarkArea(b *testing.B) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		b.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	// Create a request to pass to handler with query parameters
	req, err := inst.NewRequest("GET", "/area", nil)
	req.Form = url.Values{"lat1": {"55.660797"}, "lng1": {"12.5896"}, "lat2": {"55.663369"}, "lng2": {"12.584670"}}
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	//handler := http.Handler(appHandler(area))

	// Note that we repeat the operation we want to benchmark (in this case service.area) b.N times.
	// This value will be changed by go test until the resulting times are statistically significant.
	for i := 0; i < b.N; i++ {
		//handler.ServeHTTP(rr, req)
		area(rr, req)
	}

}

// Benchmark geo query that returns all image links within country
func BenchmarkGeo(b *testing.B) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		b.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	// Create request to pass to geo handler with query parameters
	req, err := inst.NewRequest("GET", "/area", nil)
	req.Form = url.Values{"country": {"Denmark"}, "continent": {"europe"}}
	if err != nil {
		b.Fatalf("Failed to create request: %v", err)
	}

	// Create ResponseRecorder (satisfies http.ResponseWriter) to record response
	rr := httptest.NewRecorder()

	// Run benchmark b.N times until resulting times are statistically significant
	for i := 0; i < b.N; i++ {
		geo(rr, req)
	}
}
