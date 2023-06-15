package httplayer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

func validateUpdateValueHandlersRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		w.Header().Set("Content-Type", "text/plain")
		urlPathWithoutFirstSlash := strings.TrimLeft(r.URL.Path, "/")
		pathSlice := strings.Split(urlPathWithoutFirstSlash, "/")
		if pathSlice[0] == handlePathUpdate && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if pathSlice[0] == handlePathValue && r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if pathSlice[1] != "gauge" && pathSlice[1] != "counter" {
			//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricType)
			http.Error(w, "wrong metric type", http.StatusBadRequest)
			return
		}
		if len(pathSlice) == 3 && pathSlice[2] == "" || len(pathSlice) == 2 {
			//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricName)
			http.Error(w, "wrong metric name", http.StatusNotFound)
			return
		}

		if pathSlice[0] == handlePathUpdate && len(pathSlice) < 4 {
			fmt.Println(pathSlice[0], len(pathSlice))
			http.Error(w, "wrong metric value", http.StatusBadRequest)
			return
			//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricName)
		}
		ctxWithMetricType := context.WithValue(r.Context(), contextKeyMetricType, pathSlice[1])
		ctxWithMetricName := context.WithValue(ctxWithMetricType, contextKeyMetricName, pathSlice[2])
		rWithParsedData := r.WithContext(ctxWithMetricName)
		if len(pathSlice) >= 4 {
			ctxWithMetricValue := context.WithValue(ctxWithMetricName, contextKeyMetricValue, pathSlice[3])
			rWithParsedData = r.WithContext(ctxWithMetricValue)
		}

		next.ServeHTTP(w, rWithParsedData)
	})
}
