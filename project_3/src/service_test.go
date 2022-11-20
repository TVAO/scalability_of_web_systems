// Package satservice : this contains a few integration tests of both the image and spatial area request query
package satservice

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"google.golang.org/appengine/aetest"
)

// Integration test, testing actual retrieval of images granules based on invalid lat/lng, should return error
func TestImageHandler_BadRequest(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	// Create a request to pass to handler with query parameters
	req, err := inst.NewRequest("GET", "/images", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.Handler(appHandler(images))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	//Check the response body is what we expect.
	expected := "Please provide a valid latitude and longitude"
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expected) {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'",
			rr.Body.String(), expected)
	}

}

// Integration test, testing actual retrieval of images granules based on valid lat/lng
func TestImageHandler_ValidRequest(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	// Create a request to pass to handler with query parameters
	req, err := inst.NewRequest("GET", "/images", nil)
	req.Form = url.Values{"lat": {"55.660797"}, "lng": {"12.5896"}}
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.Handler(appHandler(images))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// Integration test, testing actual retrieval of images in geographic area of interest
func TestAreaHandler_ValidRequest(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	// Create a request to pass to handler with query parameters
	req, err := inst.NewRequest("GET", "/area", nil)
	req.Form = url.Values{"lat1": {"55.660797"}, "lng1": {"12.5896"}, "lat2": {"55.663369"}, "lng2": {"12.584670"}}
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.Handler(appHandler(area))

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
