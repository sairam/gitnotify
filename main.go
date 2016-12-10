package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	yaml "gopkg.in/yaml.v2"
)

// AppConfig is
type AppConfig struct {
	ServerProto       string `yaml:"serverProto"`       // can be http:// or https://
	ServerHost        string `yaml:"serverHost"`        // domain.com with port . Used at redirection for OAuth
	LocalHost         string `yaml:"localHost"`         // host:port combination used for starting the server
	DataDir           string `yaml:"dataDir"`           // relative path from server to write the data
	SettingsFile      string `yaml:"settingsFile"`      // name of file to be looked up/saved to for data
	FromName          string `yaml:"fromName"`          // name of from email user
	FromEmail         string `yaml:"fromEmail"`         // email address of from email address
	GithubAPIEndPoint string `yaml:"githubAPIEndPoint"` // server endpoint with protocol for https://api.github.com
	GithubURLEndPoint string `yaml:"githubURLEndPoint"` // website end point https://github.com
	SMTPHost          string `yaml:"smtpHost"`
	SMTPPort          int    `yaml:"smtpPort"`
	SMTPSesConfSet    string `yaml:"sesConfigurationSet"` // ses configuration set used as a custom header while sending email
	GoogleAnalytics   string `yaml:"googleAnalytics"`
	SMTPUser          string // environment variable
	SMTPPass          string // environment variable
	RunMode           string `yaml:"runMode"` // when runMode is "dev", we use it to reload templates on every request. else they are loaded only once
}

var config = new(AppConfig)

const configFile = "config.yml"

//Page has all information about the page
type Page struct {
	Title        string
	PageTitle    string
	User         *Authentication
	Flashes      []string
	Context      interface{}
	Data         interface{}
	ClientConfig map[string]string
}

func newPage(hc *httpContext, title string, pageTitle string, conf interface{}, data interface{}) *Page {
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
		Data:      data,
	}

	page.ClientConfig = make(map[string]string)
	page.ClientConfig["GoogleAnalytics"] = config.GoogleAnalytics

	return page
}

func init() {
	loadConfig()
	go mailDaemon()
	initCron()
}

var (
	githubRepoEndPoint       string
	githubTreeURLEndPoint    string
	githubCommitURLEndPoint  string
	githubCompareURLEndPoint string
)

func loadConfig() {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		panic(err)
	}

	data, err := ioutil.ReadFile(configFile)
	if os.IsNotExist(err) {
		panic(err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}

	// load constants
	config.SMTPUser = os.Getenv("SMTP_USER")
	config.SMTPPass = os.Getenv("SMTP_PASS")

	if config.SMTPUser == "" {
		panic("Missing Configuration: SMTP username is not set!")
	}
	if config.SMTPPass == "" {
		panic("Missing Configuration: SMTP password is not set!")
	}

	githubRepoEndPoint = config.GithubURLEndPoint + "%s/"                      // repo/abc
	githubTreeURLEndPoint = config.GithubURLEndPoint + "%s/tree/%s"            // repo/abc , develop
	githubCommitURLEndPoint = config.GithubURLEndPoint + "%s/commits/%s"       // repo/abc , develop
	githubCompareURLEndPoint = config.GithubURLEndPoint + "%s/compare/%s...%s" // repo/abc, base, target commit ref

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
	r.HandleFunc("/run", forceRunHandler).Methods("POST")

	r.HandleFunc("/user", userSettingsShowHandler).Methods("GET")
	r.HandleFunc("/user", userSettingsSaveHandler).Methods("POST")

	r.HandleFunc("/typeahead/repo", repoTypeAheadHandler).Methods("GET")
	r.HandleFunc("/typeahead/branch", branchTypeAheadHandler).Methods("GET")
	r.HandleFunc("/typeahead/tz", timezoneTypeAheadHandler).Methods("GET")

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
		Addr:         config.LocalHost,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())

}
