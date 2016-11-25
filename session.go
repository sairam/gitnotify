package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewFilesystemStore("./sessions", []byte("something-very-secret"))
var sessionName = "session-name"

// responsible for setting information about the logged in user via github into the session
func setSession(r *http.Request, w http.ResponseWriter, auth, userName, token string) {

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := store.Get(r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set some session values.
	session.Values["auth"] = auth
	session.Values["user"] = userName
	session.Values["token"] = token

	// Save it before we write to the response/return from the handler.
	session.Save(r, w)

}
func userLoggedinInfo(r *http.Request) {
	session, _ := store.Get(r, sessionName)
	fmt.Println(session.Values["auth"])
	fmt.Println(session.Values["user"])
	fmt.Println(session.Values["token"])
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
	// return session
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	session.Values["auth"] = nil
	session.Values["user"] = nil
	session.Values["token"] = nil

	session.Save(r, w)

}

func isUserLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, sessionName)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	return session.Values["auth"] != nil

}
