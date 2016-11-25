package main

import "net/http"

// Responsible for displaying about page and link to login etc.,
func homeHandler(res http.ResponseWriter, req *http.Request) {
	if isUserLoggedIn(req) {
		userLoggedinInfo(req)
	}
	displayPage(res, "home", struct{}{})
}
