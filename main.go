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

//Page has all information about the page
type Page struct {
	Title     string
	PageTitle string
	User      *userInfoSession
	Flashes   []string
	Context   interface{}
}

func newPage(hc *httpContext, title string, pageTitle string, conf interface{}) *Page {
	var userInfo *userInfoSession
	if hc.isUserLoggedIn() {
		userInfo = hc.userLoggedinInfo()
	} else {
		userInfo = &userInfoSession{}
	}

	page := &Page{
		Title:     title,
		PageTitle: pageTitle,
		User:      userInfo,
		Flashes:   hc.getFlashes(),
		Context:   conf,
	}
	return page
}

// 1. Make a router to redirect user if not logged into website
// 2. Redirect to /home
// 3. Display /home page with content from tmpl/home.html
// 4. Users once logged in user goes /
// 5. Display partial to add new repository to track
// 6. Allow to autofill branch names from remote URL
func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", settingsShowHandler).Methods("GET")
	// POST is responsible for create, update and delete
	r.HandleFunc("/", settingsSaveHandler).Methods("POST")
	r.HandleFunc("/home", homeHandler).Methods("GET")

	r.HandleFunc("/logout", func(res http.ResponseWriter, req *http.Request) {
		hc := &httpContext{w: res, r: req}
		hc.clearSession()

		http.Redirect(res, req, homePageForNonLoggedIn, 302)
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
