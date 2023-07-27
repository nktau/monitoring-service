package app

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type memStorage struct {
	Gauge   map[string]float64
	Counter int64
	logger  *zap.Logger
	hashKey string
}

func New(logger *zap.Logger, key string) memStorage {
	return memStorage{
		logger:  logger,
		hashKey: key,
	}
}

var wg sync.WaitGroup

func (mem *memStorage) Start(serverURL string, reportInterval, pollInterval int) {
	//count := 0
	//for {
	//	if count%pollInterval == 0 {
	//		mem.GetRuntimeMetrics(int64(reportInterval))
	//	}
	//	count++
	//	if count%reportInterval == 0 {
	//		mem.SendRuntimeMetric(serverURL)
	//	}
	//	time.Sleep(1 * time.Second)
	//}

	wg.Add(2)
	chMemStorage := mem.GetRuntimeMetrics(int64(pollInterval))
	chMetrics := mem.CreateMetricsBuffer(chMemStorage, int64(reportInterval))
	err := mem.makeAndDoRequest(chMetrics, serverURL)
	if err != nil {
		mem.logger.Error("", zap.Error(err))
	}
	wg.Wait()

}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (mem *memStorage) GetRuntimeMetrics(pollInterval int64) chan memStorage {
	chRes := make(chan memStorage)
	go func() {
		defer wg.Done()
		//if mem.Counter%reportInterval == 0 {
		//	mem.Gauge = []map[string]float64{}
		//}
		//mem.Gauge = append(mem.Gauge, tmpGaugeMap)
		defer close(chRes)
		var memCounter int64 = 0
		for {
			tmpMemStorage := memStorage{}
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
			tmpGaugeMap["NumForcedGC"] = float64(rtm.NumForcedGC)
			tmpGaugeMap["TotalAlloc"] = float64(rtm.TotalAlloc)
			tmpGaugeMap["Lookups"] = float64(rtm.Lookups)
			tmpGaugeMap["RandomValue"] = rand.Float64()
			memCounter++
			tmpMemStorage.Gauge = tmpGaugeMap
			tmpMemStorage.Counter = memCounter
			chRes <- tmpMemStorage

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()
	return chRes
}

func (mem *memStorage) CreateMetricsBuffer(chIn chan memStorage, reportInterval int64) chan []Metrics {
	chRes := make(chan []Metrics)
	go func() {
		defer wg.Done()
		defer close(chRes)
		var metrics []Metrics
		for memStorageIter := range chIn {
			for metricName, metricValue := range memStorageIter.Gauge {
				metric := Metrics{
					ID:    metricName,
					MType: "gauge",
					Value: &metricValue,
				}
				metrics = append(metrics, metric)
			}
			metric := Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: &memStorageIter.Counter,
			}
			metrics = append(metrics, metric)
			chRes <- metrics
			time.Sleep(time.Duration(reportInterval) * time.Second)
		}
	}()

	return chRes
}

func (mem *memStorage) makeAndDoRequest(chMetrics chan []Metrics, serverURL string) error {
	for metrics := range chMetrics {
		requestBody, err := json.Marshal(metrics)
		if err != nil {
			mem.logger.Error("can't create request body json", zap.Error(err))
			return err
		}

		compressedRequestBody := mem.compress(requestBody)
		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/updates/", serverURL),
			compressedRequestBody)
		if err != nil {
			mem.logger.Error("can't create request",
				zap.Error(err),
				zap.String("request body: ", string(requestBody)))
			return err
		}

		if mem.hashKey != "" {
			hashSHA256 := mem.getSHA256HashString(compressedRequestBody)
			req.Header.Set("HashSHA256", hashSHA256)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			mem.logger.Error("can't send metric to the server",
				zap.Error(err),
				zap.String("request body: ", string(requestBody)))
			count := 0
			for {
				time.Sleep(time.Second)
				count++
				if count == 1 || count == 4 || count == 9 {
					res, err = http.DefaultClient.Do(req)
					if err != nil {
						if count == 9 {
							break
						}
					} else {
						err = req.Body.Close()
						if err != nil {
							mem.logger.Error("can't close req body", zap.Error(err))
							return err
						}
						err = res.Body.Close()
						if err != nil {
							mem.logger.Error("can't close res body", zap.Error(err))
							return err
						}
						break
					}
				}
			}
			return err
		}
		err = req.Body.Close()
		if err != nil {
			mem.logger.Error("can't close req body", zap.Error(err))
			return err
		}
		err = res.Body.Close()
		if err != nil {
			mem.logger.Error("can't close res body", zap.Error(err))
			return err
		}
	}
	return nil
}

func (mem *memStorage) compress(data []byte) *bytes.Buffer {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		mem.logger.Error("can't compress data", zap.Error(err))
	}
	if err := gz.Close(); err != nil {
		mem.logger.Error("can't close gzip writer", zap.Error(err))
	}
	return &b
}

func (mem *memStorage) getSHA256HashString(buffer *bytes.Buffer) string {
	h := hmac.New(sha256.New, []byte(mem.hashKey))
	_, err := h.Write(buffer.Bytes())
	if err != nil {
		mem.logger.Error("", zap.Error(err))
	}
	hashSHA256 := h.Sum(nil)
	hashSHA256String := fmt.Sprintf("%x", hashSHA256)
	return hashSHA256String
	//fmt.Println("hashSHA256String:", hashSHA256String)
}
