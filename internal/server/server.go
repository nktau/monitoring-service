package server

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var storage = NewMemStorage()

func updateHandler(w http.ResponseWriter, r *http.Request) {
	// http.Error or not?
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := url.Parse(strings.TrimPrefix(r.URL.Path, "/update/"))
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
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
		storage.ModifyGauge(pathSlice[1], value)
	}
	if pathSlice[0] == "counter" {
		value, err := strconv.ParseInt(pathSlice[2], 10, 64)
		if err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
		} else {
			storage.ModifyCounter(pathSlice[1], value)
		}
	}
}

func StartServer() {
	http.Handle("/update/", http.HandlerFunc(updateHandler))
	http.ListenAndServe("localhost:8080", nil)
}
