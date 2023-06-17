package httplayer

import (
	"errors"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"net/http"
	"strings"
)

var ErrMethodNotAllowed = errors.New("method not allowed")

func metricValueWithoutTrailingZero(metricValue float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", metricValue), "0"), ".")
}

func (api *httpAPI) parseAndValidateRequest(r *http.Request) (map[string]string, error) {
	requestUrlMap := map[string]string{}
	urlPathWithoutFirstSlash := strings.TrimLeft(r.URL.Path, "/")
	requestUrlSlice := strings.Split(urlPathWithoutFirstSlash, "/")
	if requestUrlSlice[0] == handlePathUpdate && r.Method != http.MethodPost {
		return nil, ErrMethodNotAllowed
	}
	if requestUrlSlice[0] == handlePathValue && r.Method != http.MethodGet {
		return nil, ErrMethodNotAllowed
	}
	requestUrlMap["location"] = requestUrlSlice[0]
	// check if metric type is empty:
	if len(requestUrlSlice) < 2 || requestUrlSlice[1] == "" {
		//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricType)
		return nil, applayer.ErrWrongMetricType
	}
	requestUrlMap["metricType"] = requestUrlSlice[1]
	//
	// check if metric name is empty:
	if len(requestUrlSlice) == 3 && requestUrlSlice[2] == "" || len(requestUrlSlice) == 2 {
		//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricName)
		return nil, applayer.ErrWrongMetricName

	}
	requestUrlMap["metricName"] = requestUrlSlice[2]
	// check if metric value is empty:
	if requestUrlSlice[0] == handlePathUpdate && len(requestUrlSlice) < 4 {
		fmt.Println(requestUrlSlice[0], len(requestUrlSlice))
		return nil, applayer.ErrWrongMetricValue
		//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricName)
	}
	if len(requestUrlSlice) >= 4 {
		requestUrlMap["metricValue"] = requestUrlSlice[3]
	}
	return requestUrlMap, nil
}

func (api *httpAPI) update(w http.ResponseWriter, r *http.Request) {
	requestUrlMap, err := api.parseAndValidateRequest(r)

	if err != nil {
		switch err {
		case ErrMethodNotAllowed:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusMethodNotAllowed)
			return
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		case applayer.ErrWrongMetricValue:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		default:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	}
	_, err = api.app.CommunicateWithHttpLayer(requestUrlMap)
	if err != nil {
		switch err {
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		case applayer.ErrWrongMetricValue:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		default:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("ok\n"))
}

func (api *httpAPI) value(w http.ResponseWriter, r *http.Request) {
	requestUrlMap, err := api.parseAndValidateRequest(r)

	if err != nil {
		switch err {
		case ErrMethodNotAllowed:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusMethodNotAllowed)
			return
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		default:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	}
	metricValue, err := api.app.CommunicateWithHttpLayer(requestUrlMap)
	if err != nil {
		switch err {
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		default:
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte(metricValue))
}

func (api *httpAPI) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauge, counter := api.app.GetAll()
	var s []string
	for key, value := range gauge {
		//to do: create a function
		s = append(s, fmt.Sprintf("<h3>%s: %s</h3>\n", key, metricValueWithoutTrailingZero(value)))
	}
	for key, value := range counter {
		s = append(s, fmt.Sprintf("<h3>%s: %d</h3>\n", key, value))
	}
	for _, i := range s {
		w.Write([]byte(i))
	}
}
