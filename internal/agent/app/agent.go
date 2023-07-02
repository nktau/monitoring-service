package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type memStorage struct {
	gauge   []map[string]float64
	counter int64
	logger  *zap.Logger
}

func New(logger *zap.Logger) memStorage {
	return memStorage{
		logger: logger,
	}
}

func (mem *memStorage) Start(serverURL string, reportInterval, pollInterval int) {
	count := 0
	for {
		if count%pollInterval == 0 {
			mem.GetRuntimeMetrics(int64(reportInterval))
		}
		count++
		if count%reportInterval == 0 {
			err := mem.SendRuntimeMetric(serverURL)
			if err != nil {
				fmt.Println("!!!!!        ERROR          !!!!", err)
				continue
			}
		}
		time.Sleep(1 * time.Second)
	}

}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (mem *memStorage) SendRuntimeMetric(serverURL string) error {
	for _, gauge := range mem.gauge {
		for metricName, metricValue := range gauge {
			metric := Metrics{
				ID:    metricName,
				MType: "gauge",
				Value: &metricValue,
			}
			requestBody, err := json.Marshal(metric)

			if err != nil {
				mem.logger.Error("can't create request body json", zap.Error(err))
				continue
			}
			req, err := http.Post(fmt.Sprintf("%s/update/", serverURL),
				"application/json",
				bytes.NewBuffer(requestBody),
			)
			if err != nil {
				mem.logger.Error("can't send metric to the server",
					zap.Error(err),
					zap.String("request body: ", string(requestBody)))
				continue
				req.Body.Close()

			}
			err = req.Body.Close()
			if err != nil {
				mem.logger.Error("can't close req body", zap.Error(err))
				continue
			}
		}
		metric := Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &mem.counter,
		}
		requestBody, err := json.Marshal(metric)
		if err != nil {
			mem.logger.Error("can't create request body json", zap.Error(err))
		}
		req, err := http.Post(fmt.Sprintf("%s/update/", serverURL),
			"application/json",
			bytes.NewBuffer(requestBody),
		)
		if err != nil {
			mem.logger.Error("can't send metric to the server", zap.Error(err))
			continue
		}
		req.Body.Close()
	}
	//metric := Metrics{
	//	ID:    "PollCount",
	//	MType: "counter",
	//	Delta: &mem.counter,
	//}
	//requestBody, err := json.Marshal(metric)
	//fmt.Println(string(requestBody))
	//if err != nil {
	//	mem.logger.Error("can't create request body json", zap.Error(err))
	//}
	//req, err := http.Post(fmt.Sprintf("%s/update/", serverURL),
	//	"application/json",
	//	bytes.NewBuffer(requestBody),
	//)
	//if err != nil {
	//	mem.logger.Error("can't send metric to the server", zap.Error(err))
	//	req.Body.Close()
	//	return err
	//}
	//req.Body.Close()
	return nil

}

func (mem *memStorage) GetRuntimeMetrics(reportInterval int64) {
	if mem.counter%reportInterval == 0 {
		mem.gauge = []map[string]float64{}
	}
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	tmpGaugeMap := map[string]float64{}
	tmpGaugeMap["Alloc"] = float64(rtm.Alloc)
	tmpGaugeMap["BuckHashSys"] = float64(rtm.BuckHashSys)
	tmpGaugeMap["Frees"] = float64(rtm.Frees)
	tmpGaugeMap["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	tmpGaugeMap["GCSys"] = float64(rtm.GCSys)
	tmpGaugeMap["HeapAlloc"] = float64(rtm.HeapAlloc)
	tmpGaugeMap["HeapIdle"] = float64(rtm.HeapIdle)
	tmpGaugeMap["HeapInuse"] = float64(rtm.HeapInuse)
	tmpGaugeMap["HeapObjects"] = float64(rtm.HeapObjects)
	tmpGaugeMap["HeapReleased"] = float64(rtm.HeapReleased)
	tmpGaugeMap["HeapSys"] = float64(rtm.HeapSys)
	tmpGaugeMap["LastGC"] = float64(rtm.LastGC)
	tmpGaugeMap["MCacheInuse"] = float64(rtm.MCacheInuse)
	tmpGaugeMap["MCacheSys"] = float64(rtm.MCacheSys)
	tmpGaugeMap["MSpanInuse"] = float64(rtm.MSpanInuse)
	tmpGaugeMap["MSpanSys"] = float64(rtm.MSpanSys)
	tmpGaugeMap["Mallocs"] = float64(rtm.Mallocs)
	tmpGaugeMap["NextGC"] = float64(rtm.NextGC)
	tmpGaugeMap["NumGC"] = float64(rtm.NumGC)
	tmpGaugeMap["OtherSys"] = float64(rtm.OtherSys)
	tmpGaugeMap["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	tmpGaugeMap["StackInuse"] = float64(rtm.StackInuse)
	tmpGaugeMap["StackSys"] = float64(rtm.StackSys)
	tmpGaugeMap["Sys"] = float64(rtm.Sys)

	tmpGaugeMap["RandomValue"] = rand.Float64()
	mem.gauge = append(mem.gauge, tmpGaugeMap)
	mem.counter = mem.counter + 1

}
