package httplayer

import (
	"fmt"
	"net/http"
)

// Такой вариант не проходит линтер, хотя мне он нравится больше
// internal/httplayer/middlewares.go:50:55: should not use built-in type string as key for value; define your own type to avoid collisions
//const contextKeyMetricType = "metricType"
//const contextKeyMetricName = "metricName"
//const contextKeyMetricValue = "metricValue"

type ContextKeyMetricType string
type ContextKeyMetricName string
type ContextKeyMetricValue string

const contextKeyMetricType ContextKeyMetricType = "metricType"
const contextKeyMetricName ContextKeyMetricName = "metricName"
const contextKeyMetricValue ContextKeyMetricValue = "metricValue"

func setHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Content-Type", "text/plain")
		next.ServeHTTP(w, r)
	})
}
