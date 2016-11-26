package main

import "net/http"

// Responsible for displaying about page and link to login etc.,
func homeHandler(res http.ResponseWriter, req *http.Request) {

	hc := &httpContext{w: res, r: req}

	page := newPage(hc, "Home Page", nil)
	displayPage(res, "home", page)
}
