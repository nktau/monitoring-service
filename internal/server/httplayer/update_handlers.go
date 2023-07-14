package httplayer

import (
	"encoding/json"
	"fmt"
	"github.com/nktau/monitoring-service/internal/server/applayer"
	"github.com/nktau/monitoring-service/internal/server/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (api *httpAPI) updateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reader, err := api.readBody(r)
	if err != nil {
		api.logger.Error("can't decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(reader)
	r.Body.Close()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		api.logger.Info("can't read request body", zap.Error(err))
		return
	}
	var metric Metrics
	err = json.Unmarshal(body, &metric)
	if err != nil {
		http.Error(w, "invalid json data", http.StatusBadRequest)
		api.logger.Info("get invalid json data from client",
			zap.String("data", string(body)),
			zap.Error(err),
		)
		return
	}
	// check if metric name is empty:
	if metric.ID == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
		return
	}
	// check if metric type is empty:
	if metric.MType == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
		return
	}
	if metric.Delta == nil && metric.Value == nil || metric.Delta != nil && metric.Value != nil {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricValue), http.StatusBadRequest)
		return
	}
	var errFromAppLayer error
	if metric.Delta != nil && metric.Value == nil {
		errFromAppLayer = api.app.Update(metric.MType, metric.ID, fmt.Sprintf("%d", *metric.Delta))
	} else if metric.Delta == nil && metric.Value != nil {
		errFromAppLayer = api.app.Update(metric.MType, metric.ID, utils.MetricValueWithoutTrailingZero(*metric.Value))
	}

	err = handleApplayerUpdateError(errFromAppLayer, w)
	if err != nil {
		return
	}
	responseBody, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (api *httpAPI) updatePlainText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	requestURLMap := map[string]string{}
	requestURLSlice := getRequestURLSlice(r.URL.Path)
	requestURLMap["location"] = requestURLSlice[0]
	// check if metric type is empty:
	if len(requestURLSlice) < 2 || requestURLSlice[1] == "" {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
		return
	}
	requestURLMap["metricType"] = requestURLSlice[1]
	//
	// check if metric name is empty:
	if len(requestURLSlice) == 3 && requestURLSlice[2] == "" || len(requestURLSlice) == 2 {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
		return
	}
	requestURLMap["metricName"] = requestURLSlice[2]
	// check if metric value is empty:
	if len(requestURLSlice) < 4 {
		http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricValue), http.StatusBadRequest)
		return
	}
	if len(requestURLSlice) >= 4 {
		requestURLMap["metricValue"] = requestURLSlice[3]
	}
	errFromAppLayer := api.app.Update(requestURLMap["metricType"], requestURLMap["metricName"], requestURLMap["metricValue"])
	err := handleApplayerUpdateError(errFromAppLayer, w)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

func handleApplayerUpdateError(errFromAppLayer error, w http.ResponseWriter) error {
	if errFromAppLayer != nil {
		switch errFromAppLayer {
		case applayer.ErrWrongMetricType:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
			return errFromAppLayer
		case applayer.ErrWrongMetricName:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
			return errFromAppLayer
		case applayer.ErrWrongMetricValue:
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricValue), http.StatusBadRequest)
			return errFromAppLayer
		default:
			http.Error(w, fmt.Sprintf("%v", errFromAppLayer), http.StatusInternalServerError)
			return errFromAppLayer
		}
	}
	return nil
}

func (api *httpAPI) whichOfUpdateHandlerUse(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "application/json" {
		api.updateJSON(w, r)
	} else {
		api.updatePlainText(w, r)
	}
}

func (api *httpAPI) updates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reader, err := api.readBody(r)
	if err != nil {
		api.logger.Error("can't decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(reader)
	r.Body.Close()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		api.logger.Info("can't read request body", zap.Error(err))
		return
	}
	var metrics []Metrics
	var appMetrics []applayer.Metrics
	err = json.Unmarshal(body, &metrics)

	if err != nil {
		http.Error(w, "invalid json data", http.StatusBadRequest)
		api.logger.Info("get invalid json data from client",
			zap.String("data", string(body)),
			zap.Error(err),
		)
		return
	}
	for _, metric := range metrics {
		if metric.ID == "" {
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricName), http.StatusNotFound)
			return
		}
		if metric.MType == "" {
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricType), http.StatusBadRequest)
			return
		}
		if metric.Delta == nil && metric.Value == nil || metric.Delta != nil && metric.Value != nil {
			http.Error(w, fmt.Sprintf("%v", applayer.ErrWrongMetricValue), http.StatusBadRequest)
			return
		}
		appMetrics = append(appMetrics, applayer.Metrics(metric))
	}

	errFromAppLayer := api.app.Updates(appMetrics)
	err = handleApplayerUpdateError(errFromAppLayer, w)
	if err != nil {
		return
	}
	//responseBody, err := json.Marshal(metrics)
	//if err != nil {
	//	http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	//	return
	//}
	//fmt.Printf("updates() my answer on request with body %s\n is %s\n", string(body), string(responseBody))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
