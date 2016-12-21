package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"

	yaml "gopkg.in/yaml.v2"
)

// AppConfig is
type AppConfig struct {
	ServerProto         string   `yaml:"serverProto"`       // can be http:// or https://
	ServerHost          string   `yaml:"serverHost"`        // domain.com with port . Used at redirection for OAuth
	LocalHost           string   `yaml:"localHost"`         // host:port combination used for starting the server
	DataDir             string   `yaml:"dataDir"`           // relative path from server to write the data
	SettingsFile        string   `yaml:"settingsFile"`      // name of file to be looked up/saved to for data
	FromName            string   `yaml:"fromName"`          // name of from email user
	FromEmail           string   `yaml:"fromEmail"`         // email address of from email address
	GithubAPIEndPoint   string   `yaml:"githubAPIEndPoint"` // server endpoint with protocol for https://api.github.com
	GithubURLEndPoint   string   `yaml:"githubURLEndPoint"` // website end point https://github.com
	GitlabAPIEndPoint   string   `yaml:"gitlabAPIEndPoint"` // server endpoint with protocol for https://gitlab.com/api/v3/
	GitlabURLEndPoint   string   `yaml:"gitlabURLEndPoint"` // website end point https://gitlab.com
	SMTPHost            string   `yaml:"smtpHost"`
	SMTPPort            int      `yaml:"smtpPort"`
	SMTPSesConfSet      string   `yaml:"sesConfigurationSet"` // ses configuration set used as a custom header while sending email
	GoogleAnalytics     string   `yaml:"googleAnalytics"`
	SMTPUser            string   // environment variable
	SMTPPass            string   // environment variable
	RunMode             string   `yaml:"runMode"` // when runMode is "dev", we use it to reload templates on every request. else they are loaded only once
	WebhookIntegrations []string `yaml:"webhookIntegrations"`
	SentryURL           string   `yaml:"sentryDSN"`

	TemplateDir         string `yaml:"templateDir"`         // tmpl/
	TemplatePartialsDir string `yaml:"templatePartialsDir"` // tmpl/partials/
	// "changes_mail" and "changes_mail_text" are the files used to render
	// "home" for home page, "repos" for the repositories page
	// "text" for rendering simple text
	// "user" for user preferences
	// use "partial name" to render a file

	Providers map[string]string // List of ProviderNames that are configured as per auth

	SourceCodeLink string
}

var config = new(AppConfig)

const (
	configFile       = "config.yml"
	runModeDev       = "dev"
	gitRefBranch     = "branches"
	gitRefTag        = "tags"
	formUpdateString = "update"
)

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
	go initCron()
}

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

	// TODO dont send email, but start the server but not the email daemon
	if config.SMTPUser == "" {
		panic("Missing Configuration: SMTP username is not set!")
	}
	if config.SMTPPass == "" {
		panic("Missing Configuration: SMTP password is not set!")
	}

	preInitAuth()

	if config.Providers["github"] != "" {
		githubRepoEndPoint = config.GithubURLEndPoint + "%s/"                      // repo/abc
		githubTreeURLEndPoint = config.GithubURLEndPoint + "%s/tree/%s"            // repo/abc , develop
		githubCommitURLEndPoint = config.GithubURLEndPoint + "%s/commits/%s"       // repo/abc , develop
		githubCompareURLEndPoint = config.GithubURLEndPoint + "%s/compare/%s...%s" // repo/abc, base, target commit ref
	}

	if config.Providers["gitlab"] != "" {
		gitlabRepoEndPoint = config.GitlabURLEndPoint + "%s/"                      // repo/abc
		gitlabTreeURLEndPoint = config.GitlabURLEndPoint + "%s/tree/%s"            // repo/abc , develop
		gitlabCommitURLEndPoint = config.GitlabURLEndPoint + "%s/commits/%s"       // repo/abc , develop
		gitlabCompareURLEndPoint = config.GitlabURLEndPoint + "%s/compare/%s...%s" // repo/abc, base, target commit ref
	}

	config.SourceCodeLink = "https://github.com/sairam/gitnotify/"

}

// 1. Make a router to redirect user if not logged into website
// 2. Redirect to /home
// 3. Display /home page with content from tmpl/home.html
// 4. Users once logged in user goes /
// 5. Display partial to add new repository to track
// 6. Allow to autofill branch names from remote URL
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

	r.HandleFunc("/logout", func(res http.ResponseWriter, req *http.Request) {
		hc := &httpContext{w: res, r: req}
		hc.clearSession()

		http.Redirect(res, req, homePageForNonLoggedIn, 302)
	}).Methods("GET")

	r.HandleFunc("/home", func(res http.ResponseWriter, req *http.Request) {
		hc := &httpContext{w: res, r: req}
		page := newPage(hc, "Home Page", "Git Notify", nil, nil)
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

// TODO - add suggestions for typeahead in firefox format
// <Url type="application/x-suggestions+json" method="GET" template="{{.Host}}/typeahead/repo?provider=github&amp;repo={searchTerms}"/>
// http://suggestqueries.google.com/complete/search?output=firefox&q=test
// output should be of the format ["test",["test match","test ranking","test","testbook","test cricket ranking","test rankings 2016"]]
// Chrome sends in a typeahead for each character pressed

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
</OpenSearchDescription>
  `
