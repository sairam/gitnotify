package main

import (
	"net/http"
	"os"
	"reflect"

	"github.com/gorilla/sessions"
)

var store = sessions.NewFilesystemStore("./sessions", []byte(os.Getenv("SESSION_FS_STORE")))
var sessionName = "_git_notify" // can be seen in the cookies list

const (
	homePageForNonLoggedIn = "/home"
	homePageForLoggedIn    = "/"
)

type httpContext struct {
	w http.ResponseWriter
	r *http.Request
}

var sv = map[string]string{
	"provider": "provider",
	"name":     "name",
	"email":    "email",
	"user":     "user",
	"token":    "token",
}

// returns true if redirected
func (hc *httpContext) redirectUnlessLoggedIn() bool {
	if !hc.isUserLoggedIn() {
		session := hc.getSession()
		if session == nil {
			return true
		}
		session.AddFlash("Login to customize settings")
		hc.saveSession(session)

		http.Redirect(hc.w, hc.r, homePageForNonLoggedIn, 302)
		return true
	}
	return false
}

// Helper methods for the session handlers
func (hc *httpContext) getSession() *sessions.Session {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := store.Get(hc.r, sessionName)
	if err != nil {
		http.Error(hc.w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return session
}

func (hc *httpContext) saveSession(session *sessions.Session) {
	// Save it before we write to the response/return from the handler.
	session.Save(hc.r, hc.w)
}

func (hc *httpContext) clearFlashes() {
	session := hc.getSession()
	if flashes := session.Flashes(); len(flashes) > 0 {
		hc.saveSession(session)
	}
}

// this actually flushes as well
func (hc *httpContext) getFlashes() []string {
	session := hc.getSession()
	if flashes := session.Flashes(); len(flashes) > 0 {
		fs := make([]string, len(flashes))
		for i, f := range flashes {
			fs[i] = string(reflect.ValueOf(f).String())
		}
		hc.saveSession(session)
		return fs
	}
	return nil
}

func (hc *httpContext) addFlash(flash string) {

	session := hc.getSession()
	if session == nil {
		return
	}

	session.AddFlash(flash)
	hc.saveSession(session)
}

// responsible for setting information about the logged in user via github into the session
func (hc *httpContext) setSession(userInfo *Authentication) {

	session := hc.getSession()
	if session == nil {
		return
	}

	session.Values[sv["provider"]] = userInfo.Provider
	session.Values[sv["name"]] = userInfo.Name
	session.Values[sv["email"]] = userInfo.Email
	session.Values[sv["user"]] = userInfo.UserName
	session.Values[sv["token"]] = userInfo.Token
	hc.clearFlashes()
	session.AddFlash("Logged in via " + userInfo.Provider)

	hc.saveSession(session)
}

func (hc *httpContext) userLoggedinInfo() *Authentication {
	userInfo := new(Authentication)

	session := hc.getSession()
	if session == nil {
		return userInfo
	}

	userInfo.Provider = reflect.ValueOf(session.Values[sv["provider"]]).String()
	userInfo.Name = reflect.ValueOf(session.Values[sv["name"]]).String()
	userInfo.Email = reflect.ValueOf(session.Values[sv["email"]]).String()
	userInfo.UserName = reflect.ValueOf(session.Values[sv["user"]]).String()
	userInfo.Token = reflect.ValueOf(session.Values[sv["token"]]).String()

	return userInfo
}

// used for logout by provider
func (hc *httpContext) clearSession() {
	session := hc.getSession()
	if session == nil {
		return
	}

	session.Values[sv["provider"]] = nil
	session.Values[sv["name"]] = nil
	session.Values[sv["email"]] = nil
	session.Values[sv["user"]] = nil
	session.Values[sv["token"]] = nil

	hc.saveSession(session)
}

func (hc *httpContext) isUserLoggedIn() bool {
	session := hc.getSession()
	return session.Values[sv["provider"]] != nil
}
