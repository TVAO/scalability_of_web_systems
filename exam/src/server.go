package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
)

// Data

type User struct {
	Name string `json:"name"`
}

var users []User

// Handlers (controllers)

// Look up user via query parameter: GET localhost:8080/users?name=Thor
func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userExists := false
	userIDQueryParam := r.FormValue("name")
	for _, user := range users {
		if user.Name == userIDQueryParam {
			userExists = true
		}
	}
	if err := json.NewEncoder(w).Encode(userExists); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//fmt.Fprintf(w, "User %s exists %s", user.Name, userExists)
	log.Printf("User exists: %s", userExists)
}

// Get all users: GET localhost:8080/users
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Post user data via request body: POST localhost:8080/users { "name": "Ole" }
func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user User
	decoder := json.NewDecoder(r.Body) // Post user data in request body
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	users = append(users, user)
	json.NewEncoder(w).Encode(users)
	log.Printf("User successfully added!")

}

// HTTP handler redirects requests to respective CRUD handlers
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.FormValue("name") == "" {
			getUsers(w, r) // Get all resources
		} else {
			get(w, r) // Serve resource
		}
	case "POST":
		post(w, r) // Create new record
	default:
		fmt.Fprintf(w, "Hello %q", html.EscapeString(r.URL.Path))
	}
}

// Uniform Routing to unique "/users" URL
func main() {
	users = append(users, User{"Thor"})
	http.HandleFunc("/users", handler)
	http.ListenAndServe(":8080", nil)
}
