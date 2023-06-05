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

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	u, err := url.Parse(strings.TrimPrefix(r.URL.Path, "/update/"))
	if err != nil {
		log.Fatal(err)
		return
	}
	// Нужна ли валидация на Content-Type == text/plain?
	pathSlice := strings.Split(u.Path, "/")
	switch len(pathSlice) {
	case 0:
		http.Error(w, "wrong metric type", http.StatusBadRequest)
		return
	case 1:
		http.Error(w, "wrong metric name", http.StatusNotFound)
		return
	case 2:
		http.Error(w, "wrong metric value", http.StatusBadRequest)
		return
	}
	if pathSlice[0] != "gauge" && pathSlice[0] != "counter" {
		http.Error(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	if pathSlice[1] == "gauge" || pathSlice[1] == "counter" || pathSlice[1] == "" {
		http.Error(w, "wrong metric name", http.StatusNotFound)
		return
	}
	if pathSlice[0] == "gauge" {
		if value, err := strconv.ParseFloat(pathSlice[2], 64); err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
		} else {
			storage.ModifyGauge(pathSlice[1], value)
		}
	} else if pathSlice[0] == "counter" {
		if value, err := strconv.ParseInt(pathSlice[2], 10, 64); err != nil {
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
