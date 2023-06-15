package app

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type memStorage struct {
	gauge   []map[string]float64
	counter int64
}

func New() memStorage {
	return memStorage{}
}

func (mem *memStorage) Start(serverURL string, reportInterval, pollInterval int) {
	count := 0
	for {
		if count%pollInterval == 0 {
			mem.GetRuntimeMetrics()
		}
		count++
		if count%reportInterval == 0 {
			mem.SendRuntimeMetric(serverURL)
		}
		time.Sleep(1 * time.Second)
	}

}

func (mem *memStorage) SendRuntimeMetric(serverURL string) {
	for _, gauge := range mem.gauge {
		for metricName, metricValue := range gauge {
			requestURL := fmt.Sprintf("%s/update/gauge/%s/%f", serverURL, metricName, metricValue)
			req, err := http.Post(requestURL, "text/plain", nil)
			if err != nil {
				fmt.Println(err)
				req.Body.Close()
			}
			req.Body.Close()
		}
	}
}

func (mem *memStorage) GetRuntimeMetrics() {
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