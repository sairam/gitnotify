package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sairam/timezone"
)

var tzByOffset map[int][]timezone.Timezone

func init() {
	tzByOffset = timezone.GroupLocationByOffset()
}

func timezoneTypeAheadHandler(w http.ResponseWriter, r *http.Request) {
	var offset int
	if len(r.URL.Query()["offset"]) > 0 {
		inputOffset, err := strconv.ParseFloat(r.URL.Query()["offset"][0], 10)

		if err != nil {
			http.NotFound(w, r)
			return
		}
		// if offset is of the format 5.5, answer is 5.5*3600
		// if offset is of the format +0530 - convert 5.5 and then use above
		offset = int(inputOffset * 3600)

	}

	// allows jsonp via "callback"
	callback := ""
	if len(r.URL.Query()["callback"]) > 0 {
		callback = r.URL.Query()["callback"][0]
	}

	var data []timezone.Timezone
	if len(r.URL.Query()["offset"]) > 0 {
		data = tzByOffset[offset]
	} else {
		data = timezone.Locations
	}

	options := make([]string, 0, len(data))
	for _, tz := range data {
		options = append(options, tz.Location)
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(options)

	cacheSince := time.Now().Format(http.TimeFormat)
	// cache for 1 year
	cacheUntil := time.Now().AddDate(1, 0, 0).Format(http.TimeFormat)
	maxAge := time.Now().AddDate(1, 0, 0).Unix()
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age:%d, public", maxAge))
	w.Header().Set("Last-Modified", cacheSince)
	w.Header().Set("Expires", cacheUntil)
	w.Header().Set("Content-Type", "application/json")

	if callback != "" {
		w.Write([]byte(callback + "("))
	}
	io.Copy(w, b)
	if callback != "" {
		w.Write([]byte(")"))
	}

}
