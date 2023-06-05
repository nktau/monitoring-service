package main

import (
	"fmt"
	"github.com/nktau/monitoring-service/internal/server"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var memStorage = server.NewMemStorage()

func updateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	u, err := url.Parse(strings.TrimPrefix(r.URL.Path, "/update/"))
	if err != nil {
		log.Fatal(err)
	}
	// Нужна ли валидация на Content-Type == text/plain?
	pathSlice := strings.Split(u.Path, "/")
	fmt.Println(pathSlice)
	if pathSlice[0] != "gauge" && pathSlice[0] != "counter" {
		http.Error(w, "wrong metric type", http.StatusBadRequest)
	}
	if pathSlice[1] == "gauge" || pathSlice[1] == "counter" || pathSlice[1] == "" {
		http.Error(w, "wrong metric name", http.StatusNotFound)
	}
	if pathSlice[0] == "gauge" {
		if value, err := strconv.ParseFloat(pathSlice[2], 64); err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
		} else {
			memStorage.ModifyGauge(pathSlice[1], value)
		}
	} else if pathSlice[0] == "counter" {
		if value, err := strconv.ParseInt(pathSlice[2], 10, 64); err != nil {
			http.Error(w, "wrong metric value", http.StatusBadRequest)
		} else {
			memStorage.ModifyCounter(pathSlice[1], value)
		}
	}
}

func main() {

	http.Handle("/update/", http.HandlerFunc(updateHandler))
	http.ListenAndServe("localhost:8080", nil)

}
