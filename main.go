package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

const (
	appConfigFile    = "config.yml"
	runModeDev       = "dev"
	gitRefBranch     = "branches"
	gitRefTag        = "tags"
	formUpdateString = "update"
)

var (
	githubRepoEndPoint       string
	githubTreeURLEndPoint    string
	githubCommitURLEndPoint  string
	githubCompareURLEndPoint string

	gitlabRepoEndPoint       string
	gitlabTreeURLEndPoint    string
	gitlabCommitURLEndPoint  string
	gitlabCompareURLEndPoint string
)

func init() {
	loadConfig()
	if config.isEmailSetup() {
		go mailDaemon()
	} else {
		log.Print("Email is not configured")
	}
	go initCron()
}

func main() {
	r := mux.NewRouter()
	auth := r.PathPrefix("/auth").Subrouter()
	initAuth(auth)

	r.HandleFunc("/", settingsShowHandler).Methods("GET")
	// POST is responsible for create, update and delete
	r.HandleFunc("/", settingsSaveHandler).Methods("POST")
	r.HandleFunc("/run", forceRunHandler).Methods("POST")

	r.HandleFunc("/user", userSettingsShowHandler).Methods("GET")
	r.HandleFunc("/user", userSettingsSaveHandler).Methods("POST")

	r.HandleFunc("/typeahead/repo", repoTypeAheadHandler).Methods("GET")
	r.HandleFunc("/typeahead/branch", branchTypeAheadHandler).Methods("GET")
	r.HandleFunc("/typeahead/tz", timezoneTypeAheadHandler).Methods("GET")

	r.HandleFunc("/changes", listAllDiffs).Methods("GET")
	r.HandleFunc("/changes/", listAllDiffs).Methods("GET")
	r.HandleFunc("/changes/{diffentry}", renderThisDiff).Methods("GET")

	r.HandleFunc("/logout", func(res http.ResponseWriter, req *http.Request) {
		hc := &httpContext{w: res, r: req}
		hc.clearSession()

		http.Redirect(res, req, homePageForNonLoggedIn, 302)
	}).Methods("GET")

	r.HandleFunc("/home", func(res http.ResponseWriter, req *http.Request) {
		hc := &httpContext{w: res, r: req}
		page := newPage(hc, "Home Page", "Get Daily Code Diffs from Repositories", nil, nil)
		displayPage(res, "home", page)
	}).Methods("GET")

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "robots.txt", time.Now(), strings.NewReader("User-agent: *\n"))
	})

	r.HandleFunc("/opensearch.xml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		t, _ := template.New("foo").Parse(opensearchTemplate)
		t.Execute(w, &struct{ Host string }{config.ServerProto + config.ServerHost})
	})

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		d := fmt.Sprintf("status:%s\ntime:%s\n", "pong", time.Now())
		http.ServeContent(w, r, "ping", time.Now(), strings.NewReader(d))
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         config.LocalHost,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}

var opensearchTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/" xmlns:moz="http://www.mozilla.org/2006/browser/search/">
  <ShortName>gitnotify</ShortName>
  <Description>Track Code Diffs from Github &amp; Gitlab</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Tags>git notify gitnotify</Tags>
  <Image height="16" width="16" type="image/x-icon">{{.Host}}/favicon.ico</Image>
  <SearchForm>{{.Host}}</SearchForm>
  <Url type="text/html" method="GET" template="{{.Host}}/?repo={searchTerms}&amp;utm_source=opensearch" />
  <Query role="example" searchTerms="rails/rails" />
</OpenSearchDescription>`
