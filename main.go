package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	serverProto = "http://"
	host        = "localhost:3000"
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
	r.HandleFunc("/", settingsShowHandler).Methods("GET")
	r.HandleFunc("/", settingsSaveHandler).Methods("POST")
	r.HandleFunc("/home", homeHandler).Methods("GET")

	r.HandleFunc("/logout", func(res http.ResponseWriter, req *http.Request) {
		clearSession(res, req)
		// redirect to /home
	}).Methods("GET")

	auth := r.PathPrefix("/auth").Subrouter()
	initAuth(auth)

	srv := &http.Server{
		Handler:      r,
		Addr:         host,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())

}
