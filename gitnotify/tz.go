package gitnotify

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sairam/timezone"
)

var tzByOffset map[int][]timezone.Timezone

func initTZ() {
	go func() { tzByOffset = timezone.GroupLocationByOffset() }()
}

func timezoneTypeAheadHandler(w http.ResponseWriter, r *http.Request) {
	var offset int
	var offsetStr string
	cacher, setCache := w.(CacheWriterIface)
	var cacheResponse = config.CacheMode

	// allows jsonp via "callback"
	var callback = ""
	if len(r.URL.Query()["callback"]) > 0 {
		callback = r.URL.Query()["callback"][0]
		cacheResponse = false
	}

	if len(r.URL.Query()["offset"]) > 0 {
		offsetStr = r.URL.Query()["offset"][0]
		inputOffset, err := strconv.ParseFloat(offsetStr, 10)

		if err != nil {
			http.NotFound(w, r)
			return
		}
		if cacheResponse && setCache {
			cacher.SetCachePath("tz-" + offsetStr)
			if cacher.WriteFromCache() {
				return
			}
		}

		// if offset is of the format 5.5, answer is 5.5*3600
		// if offset is of the format +0530 - convert 5.5 and then use above
		offset = int(inputOffset * 3600)
	}

	var data []timezone.Timezone
	if len(offsetStr) > 0 {
		data = tzByOffset[offset]
	} else {
		data = timezone.Locations
	}

	options := make([]string, 0, len(data))
	for _, tz := range data {
		options = append(options, tz.Location)
	}

	// set to end of this year instead of 1 day
	if cacheResponse {
		setCacheHeaders(w)
	}
	// w.Header().Set("Content-Type", "application/json")

	if callback != "" {
		w.Write([]byte(callback + "("))
	}
	json.NewEncoder(w).Encode(options)
	if callback != "" {
		w.Write([]byte(")"))
	}

}
