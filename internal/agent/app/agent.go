package app

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	psLoad "github.com/shirou/gopsutil/v3/load"
	psMem "github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type memStorage struct {
	Gauge   map[string]float64
	Counter int64
}

type agent struct {
	data   memStorage
	logger *zap.Logger
	config MemStorageConfig
}

type MemStorageConfig struct {
	ServerURL      string
	ReportInterval int
	PollInterval   int
	HashKey        string
	RateLimit      int
}

func New(logger *zap.Logger, memStorageConfig MemStorageConfig) agent {
	return agent{
		logger: logger,
		config: memStorageConfig,
	}
}

var wg sync.WaitGroup

func (mem *agent) Start() {
	wg.Add(2)
	chRuntimeMetrics := mem.GetRuntimeMetrics()
	chGopsutilMetrics := mem.GetGopsutilMetrics()
	chMemStorage := mem.CombineGettingMetrics(chRuntimeMetrics, chGopsutilMetrics)
	chMetrics := mem.CreateMetricsBuffer(chMemStorage)
	g := new(errgroup.Group)
	for i := 0; i < mem.config.RateLimit; i++ {
		g.Go(func() error {
			return mem.makeAndDoRequest(chMetrics)
		})
	}
	if err := g.Wait(); err != nil {
		mem.logger.Error("", zap.Error(err))
	}
	wg.Wait()
}

type Metrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (mem *agent) CombineGettingMetrics(chRuntime, chGopsutil chan map[string]float64) chan memStorage {
	chRes := make(chan memStorage)
	go func() {
		var counter int64 = 0
		for {
			tmpMemStorage := memStorage{}
			tmpGaugeMap := <-chRuntime
			for key, value := range <-chGopsutil {
				tmpGaugeMap[key] = value
			}
			tmpMemStorage.Gauge = tmpGaugeMap
			tmpMemStorage.Counter = counter
			chRes <- tmpMemStorage
			counter++
		}
	}()

	return chRes
}

func (mem *agent) GetRuntimeMetrics() chan map[string]float64 {
	chRes := make(chan map[string]float64)
	go func() {
		defer wg.Done()
		defer close(chRes)
		for {
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
			chRes <- tmpGaugeMap
			time.Sleep(time.Duration(mem.config.PollInterval) * time.Second)
		}
	}()
	return chRes
}

func (mem *agent) GetGopsutilMetrics() chan map[string]float64 {
	chRes := make(chan map[string]float64)
	go func() {
		defer close(chRes)
		for {
			memoryUsage, err := psMem.VirtualMemory()
			if err != nil {
				mem.logger.Error("", zap.Error(err))
			}
			tmpGaugeMap := map[string]float64{}
			tmpGaugeMap["TotalMemory"] = float64(memoryUsage.Total)
			tmpGaugeMap["FreeMemory"] = float64(memoryUsage.Free)
			load, err := psLoad.Avg()
			if err != nil {
				mem.logger.Error("", zap.Error(err))
			}
			tmpGaugeMap["CPUutilization1"] = load.Load1
			chRes <- tmpGaugeMap
			time.Sleep(time.Duration(mem.config.PollInterval) * time.Second)
		}
	}()

	return chRes
}

func (mem *agent) CreateMetricsBuffer(chIn chan memStorage) chan []Metrics {
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
					Value: metricValue,
				}
				metrics = append(metrics, metric)
			}
			metric := Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: memStorageIter.Counter,
			}
			metrics = append(metrics, metric)
			chRes <- metrics
			time.Sleep(time.Duration(mem.config.ReportInterval) * time.Second)
		}
	}()
	return chRes
}

func (mem *agent) makeAndDoRequest(chMetrics chan []Metrics) error {
	for metrics := range chMetrics {
		requestBody, err := json.Marshal(metrics)
		if err != nil {
			mem.logger.Error("can't create request body json", zap.Error(err))
			return err
		}
		compressedRequestBody := mem.compress(requestBody)

		for i := 0; i < 11; i++ {
			req, err := http.NewRequest(http.MethodPost,
				fmt.Sprintf("%s/updates/", mem.config.ServerURL),
				compressedRequestBody)
			if err != nil {
				mem.logger.Error("can't create request",
					zap.Error(err),
					zap.String("request body: ", string(requestBody)))
				return err
			}

			if mem.config.HashKey != "" {
				hashSHA256 := mem.getSHA256HashString(compressedRequestBody)
				req.Header.Set("HashSHA256", hashSHA256)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				headers := ""
				for header, _ := range req.Header {
					headers += fmt.Sprintf("%s: %s, ", header, req.Header.Get(header))
				}
				mem.logger.Error("can't send metric to the server",
					zap.Error(err),
					zap.String("request body: ", string(requestBody)),
					zap.String("request body: ", string(requestBody)),
					zap.String("request headers", headers),
				)
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
			time.Sleep(time.Second)

		}
	}
	return nil
}

func (mem *agent) compress(data []byte) *bytes.Buffer {
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

func (mem *agent) getSHA256HashString(buffer *bytes.Buffer) string {
	h := hmac.New(sha256.New, []byte(mem.config.HashKey))
	_, err := h.Write(buffer.Bytes())
	if err != nil {
		mem.logger.Error("", zap.Error(err))
	}
	hashSHA256 := h.Sum(nil)
	hashSHA256String := fmt.Sprintf("%x", hashSHA256)
	return hashSHA256String
}

//resBody, err := io.ReadAll(res.Body)
//mem.logger.Debug("send metrics to the server",
//	zap.String("URL", req.URL.String()),
//	zap.String("status code", res.Status),
//	zap.String("Method", req.Method),
//	zap.String("response body", string(resBody)),
//)
