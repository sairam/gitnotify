package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewFilesystemStore("./sessions", []byte("something-very-secret"))
var sessionName = "session-name"

type userInfoSession struct {
	auth     string
	userName string
	token    string
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

// responsible for setting information about the logged in user via github into the session
func (hc *httpContext) setSession(userInfo *userInfoSession) {

	session := hc.getSession()
	if session == nil {
		return
	}

	session.Values[sv["authType"]] = userInfo.auth
	session.Values[sv["user"]] = userInfo.userName
	session.Values[sv["token"]] = userInfo.token

	hc.saveSession(session)
}

func (hc *httpContext) userLoggedinInfo() {
	session := hc.getSession()
	if session == nil {
		return
	}

	fmt.Println(session.Values[sv["authType"]])
	fmt.Println(session.Values[sv["user"]])
	fmt.Println(session.Values[sv["token"]])
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
