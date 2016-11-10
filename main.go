package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// 1. Make a router to redirect user if not logged into website
// 2. Redirect to /home
// 3. Display /home page with content from tmpl/home.html
// 4. Users once logged in user goes /
// 5. Display partial to add new repository to track and information via AJAX to auto fill branch names separated with ','
func main() {

	// Session handler is created?
	// initSession()

	r := mux.NewRouter()
	r.HandleFunc("/", settingsHandler)
	r.HandleFunc("/home", homeHandler)
	auth := r.PathPrefix("/auth").Subrouter()
	initAuth(auth)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:3000",
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())

}
