package httplayer

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (api *httpAPI) update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := url.Parse(strings.TrimPrefix(r.URL.Path, "/update/"))
	if err != nil {
		log.Fatal(err)
		// to do add 500 error
		return
	}
	pathSlice := strings.Split(u.Path, "/")
	if pathSlice[0] != "gauge" && pathSlice[0] != "counter" {
		http.Error(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if len(pathSlice) == 2 && pathSlice[1] == "" || len(pathSlice) == 1 {
		http.Error(w, "wrong metric name", http.StatusNotFound)
		return
	}
	if len(pathSlice) < 3 {
		http.Error(w, "wrong metric value", http.StatusBadRequest)
		return
	}
	if pathSlice[0] == "gauge" {
		value, err := strconv.ParseFloat(pathSlice[2], 64)
		if err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
		}
		api.app.UpdateGauge(pathSlice[1], value)
	}
	if pathSlice[0] == "counter" {
		value, err := strconv.ParseInt(pathSlice[2], 10, 64)
		if err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
		}
		api.app.UpdateCounter(pathSlice[1], value)
	}
	w.Write([]byte("ok\n"))
}
