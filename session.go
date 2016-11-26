package main

import (
	"net/http"
	"reflect"

	"github.com/gorilla/sessions"
)

var store = sessions.NewFilesystemStore("./sessions", []byte("something-very-secret"))
var sessionName = "session-name"

const (
	homePageForNonLoggedIn = "/home"
	homePageForLoggedIn    = "/"
)

type userInfoSession struct {
	Auth     string
	UserName string
	Token    string
}

type httpContext struct {
	w http.ResponseWriter
	r *http.Request
}

var sv = make(map[string]string)

func init() {
	sv["authType"] = "auth"
	sv["user"] = "user"
	sv["token"] = "token"
}

// returns true if redirected
func (hc *httpContext) redirectUnlessLoggedIn() bool {
	if !hc.isUserLoggedIn() {
		session := hc.getSession()
		if session == nil {
			return true
		}
		session.AddFlash("You need to be logged in to configure")
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

// responsible for setting information about the logged in user via github into the session
func (hc *httpContext) setSession(userInfo *userInfoSession) {

	session := hc.getSession()
	if session == nil {
		return
	}

	session.Values[sv["authType"]] = userInfo.Auth
	session.Values[sv["user"]] = userInfo.UserName
	session.Values[sv["token"]] = userInfo.Token
	hc.clearFlashes()
	session.AddFlash("Logged in via " + userInfo.Auth)

	hc.saveSession(session)
}

func (hc *httpContext) userLoggedinInfo() *userInfoSession {
	userInfo := new(userInfoSession)

	session := hc.getSession()
	if session == nil {
		return userInfo
	}

	userInfo.Auth = reflect.ValueOf(session.Values[sv["authType"]]).String()
	userInfo.UserName = reflect.ValueOf(session.Values[sv["user"]]).String()
	userInfo.Token = reflect.ValueOf(session.Values[sv["token"]]).String()

	return userInfo
}

// used for logout by authType
func (hc *httpContext) clearSession() {
	session := hc.getSession()
	if session == nil {
		return
	}

	session.Values[sv["authType"]] = nil
	session.Values[sv["user"]] = nil
	session.Values[sv["token"]] = nil

	hc.saveSession(session)
}

func (hc *httpContext) isUserLoggedIn() bool {
	session := hc.getSession()
	return session.Values[sv["authType"]] != nil
}
