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
	requestURLMap := map[string]string{}
	urlPathWithoutFirstSlash := strings.TrimLeft(r.URL.Path, "/")
	requestURLSlice := strings.Split(urlPathWithoutFirstSlash, "/")
	if requestURLSlice[0] == handlePathUpdate && r.Method != http.MethodPost {
		return nil, ErrMethodNotAllowed
	}
	if requestURLSlice[0] == handlePathValue && r.Method != http.MethodGet {
		return nil, ErrMethodNotAllowed
	}
	requestURLMap["location"] = requestURLSlice[0]
	// check if metric type is empty:
	if len(requestURLSlice) < 2 || requestURLSlice[1] == "" {
		return nil, applayer.ErrWrongMetricType
	}
	requestURLMap["metricType"] = requestURLSlice[1]
	//
	// check if metric name is empty:
	if len(requestURLSlice) == 3 && requestURLSlice[2] == "" || len(requestURLSlice) == 2 {
		return nil, applayer.ErrWrongMetricName

	}
	requestURLMap["metricName"] = requestURLSlice[2]
	// check if metric value is empty:
	if requestURLSlice[0] == handlePathUpdate && len(requestURLSlice) < 4 {
		fmt.Println(requestURLSlice[0], len(requestURLSlice))
		return nil, applayer.ErrWrongMetricValue
		//ctxWithUser = context.WithValue(r.Context(), validateErrorKey, validateErrorValueWrongMetricName)
	}
	if len(requestURLSlice) >= 4 {
		requestURLMap["metricValue"] = requestURLSlice[3]
	}
	return requestURLMap, nil
}

func (api *httpAPI) update(w http.ResponseWriter, r *http.Request) {
	requestURLMap, err := api.parseAndValidateRequest(r)

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
	_, err = api.app.CommunicateWithHTTPLayer(requestURLMap)
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
	requestURLMap, err := api.parseAndValidateRequest(r)

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
	metricValue, err := api.app.CommunicateWithHTTPLayer(requestURLMap)
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
	w.Write([]byte(fmt.Sprintf("%s\n", metricValue)))
}

func (api *httpAPI) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	gauge, counter := api.app.GetAll()
	var s []string
	for key, value := range gauge {
		s = append(s, fmt.Sprintf("<h3>%s: %s</h3>\n", key, metricValueWithoutTrailingZero(value)))
	}
	for key, value := range counter {
		s = append(s, fmt.Sprintf("<h3>%s: %d</h3>\n", key, value))
	}
	for _, i := range s {
		w.Write([]byte(i))
	}
}
