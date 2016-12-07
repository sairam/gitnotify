package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

// https://github.com/markbates/goth/blob/master/providers/github/github.go
//	github.AuthURL = "https://github.acme.com/login/oauth/authorize
//	github.TokenURL = "https://github.acme.com/login/oauth/access_token
//	github.ProfileURL = "https://github.acme.com/api/v3/user

// Customizing Auth/Token/Profile URL unavailable for Gitlab.
// TODO send PR
// https://github.com/markbates/goth/blob/master/providers/gitlab/gitlab.go

// https://developer.github.com/v3/oauth/#scopes
// for github, add scope: "repo:status" to access private repositories

// data/$provider/$user/settings.yml
type Authentication struct {
	Provider string `yaml:"provider"` // github/gitlab
	Name     string `yaml:"name"`     // name of the person addressing to
	Email    string `yaml:"email"`    // email that we will send to
	UserName string `yaml:"username"` // username for identification
	Token    string `yaml:"token"`    // used to query the provider
}

func (userInfo *Authentication) save() {

	conf := new(Setting)
	conf.load(userInfo.getConfigFile())
	conf.Auth = userInfo
	conf.save(userInfo.getConfigFile())

}

func (userInfo *Authentication) getConfigFile() string {
	if userInfo.Provider == "" {
		return ""
	}
	return fmt.Sprintf("data/%s/%s/settings.yml", userInfo.Provider, userInfo.UserName)
}

// ProviderIndex is used for setting up the providers
type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

func init() {
	if os.Getenv("GITHUB_KEY") == "" || os.Getenv("GITHUB_SECRET") == "" {
		panic("Missing Configuration: Github Authentication is not set!")
	}
	gothic.Store = sessions.NewFilesystemStore(os.TempDir(), []byte("goth-example"))
	gothic.GetProviderName = getProviderName
}

// load envconfig via https://github.com/kelseyhightower/envconfig
func initAuth(p *mux.Router) {
	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), serverProto+host+"/auth/github/callback"),
		// gitlab.New(os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"), serverProto+host+"/auth/github/callback"),
	)

	m := map[string]string{
		"github": "Github",
		// "gitlab": "GitLab",
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	// p := pat.New()
	p.HandleFunc("/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		authType, _ := getProviderName(req)
		auth := &Authentication{
			Provider: authType,
			UserName: user.NickName,
			Name:     user.Name,
			Email:    user.Email,
			Token:    user.AccessToken,
		}
		auth.save()

		hc := &httpContext{res, req}
		hc.setSession(auth)

		http.Redirect(res, req, homePageForLoggedIn, 302)
	}).Methods("GET")

	p.HandleFunc("/{provider}", func(res http.ResponseWriter, req *http.Request) {
		hc := &httpContext{res, req}
		if hc.isUserLoggedIn() {
			text := "User is already logged in"
			displayText(hc, res, text)
		} else {
			gothic.BeginAuthHandler(res, req)
		}
	}).Methods("GET")

	p.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.New("foo").Parse(indexTemplate)
		t.Execute(res, providerIndex)
	}).Methods("GET")
}

// See gothic/gothic.go: GetProviderName function
// Overridden since we use mux
func getProviderName(req *http.Request) (string, error) {
	vars := mux.Vars(req)
	provider := vars["provider"]
	if provider == "" {
		return provider, errors.New("you must select a provider")
	}
	return provider, nil
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
