package app

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/go-resty/resty/v2"
	psLoad "github.com/shirou/gopsutil/v3/load"
	psMem "github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	"math/rand"
	"runtime"
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

func (mem *agent) Start() {
	chRuntimeMetrics := mem.GetRuntimeMetrics()
	chGopsutilMetrics := mem.GetGopsutilMetrics()
	chMemStorage := mem.CombineGettingMetrics(chRuntimeMetrics, chGopsutilMetrics)
	chMetrics := mem.CreateMetricsBuffer(chMemStorage)

	// Запускаем отправку запросов последовательно, без горутин
	for i := 0; i < mem.config.RateLimit; i++ {
		if err := mem.makeAndDoRequest(chMetrics); err != nil {
			mem.logger.Error("", zap.Error(err))
		}
	}
}

type Metrics struct {
	ID    string  `json:"id"`              // имя метрики
	MType string  `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (mem *agent) CombineGettingMetrics(chRuntime, chGopsutil map[string]float64) memStorage {
	var counter int64 = 0
	tmpMemStorage := memStorage{}
	tmpGaugeMap := chRuntime
	for key, value := range chGopsutil {
		tmpGaugeMap[key] = value
	}
	tmpMemStorage.Gauge = tmpGaugeMap
	tmpMemStorage.Counter = counter
	counter++
	return tmpMemStorage
}

func (mem *agent) GetRuntimeMetrics() map[string]float64 {
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
	time.Sleep(time.Duration(mem.config.PollInterval) * time.Second)
	return tmpGaugeMap
}

func (mem *agent) GetGopsutilMetrics() map[string]float64 {
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
	time.Sleep(time.Duration(mem.config.PollInterval) * time.Second)
	return tmpGaugeMap
}

func (mem *agent) CreateMetricsBuffer(memStorage memStorage) []Metrics {
	var metrics []Metrics
	for metricName, metricValue := range memStorage.Gauge {
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
		Delta: memStorage.Counter,
	}
	metrics = append(metrics, metric)
	time.Sleep(time.Duration(mem.config.ReportInterval) * time.Millisecond)
	return metrics
}

func (mem *agent) makeAndDoRequest(metrics []Metrics) error {
	client := resty.New()
	_, err := client.R().SetHeader("Content-Type", "application/json").
		SetBody(metrics).
		Post(fmt.Sprintf("%s/updates/", mem.config.ServerURL))
	if err != nil {
		mem.logger.Error("can't do request", zap.Error(err))
	}
	return err
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
