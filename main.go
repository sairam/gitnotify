package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	serverProto             = "http://"
	host                    = "localhost:3000"
	dataDir                 = "./data"
	settingsFile            = "settings.yml"
	fromName                = "Git Notify"
	fromEmail               = "sairam.kunala@gmail.com"
	githubAPIEndPoint       = "https://api.github.com/repos/"
	githubURLEndPoint       = "https://github.com/%s/"           // repo/abc
	githubTreeURLEndPoint   = "https://github.com/%s/tree/%s"    // repo/abc , develop
	githubCommitURLEndPoint = "https://github.com/%s/commits/%s" // repo/abc , develop
)

//Page has all information about the page
type Page struct {
	Title     string
	PageTitle string
	User      *Authentication
	Flashes   []string
	Context   interface{}
}

func newPage(hc *httpContext, title string, pageTitle string, conf interface{}) *Page {
	var userInfo *Authentication
	if hc.isUserLoggedIn() {
		userInfo = hc.userLoggedinInfo()
	} else {
		userInfo = &Authentication{}
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
