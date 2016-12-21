package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
)

// Authentication data/$provider/$user/$settingsFile
type Authentication struct {
	Provider string `yaml:"provider"` // github/gitlab
	Name     string `yaml:"name"`     // name of the person addressing to
	Email    string `yaml:"email"`    // email that we will send to
	UserName string `yaml:"username"` // username for identification
	Token    string `yaml:"token"`    // used to query the provider
}

func (userInfo *Authentication) save() {

	conf := new(Setting)
	os.MkdirAll(userInfo.getConfigDir(), 0700)
	conf.load(userInfo.getConfigFile())
	conf.Auth = userInfo
	conf.save(userInfo.getConfigFile())

}

func (userInfo *Authentication) getConfigDir() string {
	if userInfo.Provider == "" {
		return ""
	}
	return fmt.Sprintf("data/%s/%s", userInfo.Provider, userInfo.UserName)
}

func (userInfo *Authentication) getConfigFile() string {
	if userInfo.Provider == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", userInfo.getConfigDir(), config.SettingsFile)
}

func init() {
	gothic.Store = store
	gothic.GetProviderName = getProviderName
}

func preInitAuth() {
	// ProviderNames is the map of key/value providers configured
	config.Providers = make(map[string]string)

	var providers []goth.Provider

	if provider := configureGithub(); provider != nil {
		providers = append(providers, provider)
	}

	if provider := configureGitlab(); provider != nil {
		providers = append(providers, provider)
	}

	goth.UseProviders(providers...)
}

func initAuth(p *mux.Router) {
	p.HandleFunc("/{provider}/callback", authProviderCallbackHandler).Methods("GET")
	p.HandleFunc("/{provider}", authProviderHandler).Methods("GET")
	p.HandleFunc("/", authListHandler).Methods("GET")
}

func configureGithub() goth.Provider {
	if config.GithubURLEndPoint != "" && config.GithubAPIEndPoint != "" {
		if os.Getenv("GITHUB_KEY") == "" || os.Getenv("GITHUB_SECRET") == "" {
			panic("Missing Configuration: Github Authentication is not set!")
		}

		github.AuthURL = config.GithubURLEndPoint + "login/oauth/authorize"
		github.TokenURL = config.GithubURLEndPoint + "login/oauth/access_token"
		github.ProfileURL = config.GithubAPIEndPoint + "user"

		config.Providers["github"] = "Github"
		// for github, add scope: "repo:status" to access private repositories
		return github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), config.ServerProto+config.ServerHost+"/auth/github/callback", "user:email")
	}
	return nil
}

func configureGitlab() goth.Provider {
	if config.GitlabURLEndPoint != "" && config.GitlabAPIEndPoint != "" {
		if os.Getenv("GITLAB_KEY") == "" || os.Getenv("GITLAB_SECRET") == "" {
			panic("Missing Configuration: Github Authentication is not set!")
		}

		gitlab.AuthURL = config.GitlabURLEndPoint + "oauth/authorize"
		gitlab.TokenURL = config.GitlabURLEndPoint + "oauth/token"
		gitlab.ProfileURL = config.GitlabAPIEndPoint + "user"

		config.Providers["gitlab"] = "Gitlab"
		// gitlab does not have any scopes, you get full access to the user's account
		return gitlab.New(os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"), config.ServerProto+config.ServerHost+"/auth/gitlab/callback")
	}
	return nil

}

func authListHandler(res http.ResponseWriter, req *http.Request) {
	var keys []string
	for k := range config.Providers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: config.Providers}

	t, _ := template.New("foo").Parse(indexTemplate)
	t.Execute(res, providerIndex)
}

func authProviderHandler(res http.ResponseWriter, req *http.Request) {
	hc := &httpContext{res, req}
	if hc.isUserLoggedIn() {
		text := "User is already logged in"
		displayText(hc, res, text)
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

func authProviderCallbackHandler(res http.ResponseWriter, req *http.Request) {
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
	hc.setSession(auth, authType)

	http.Redirect(res, req, homePageForLoggedIn, 302)
}

// ProviderIndex is used for setting up the providers
type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
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
