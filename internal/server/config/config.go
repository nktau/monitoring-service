package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type config struct {
	ListenAddress   string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

var Config config

func Init() {
	Config.parseFlags()
	Config.parseEnv()
}
func (cfg *config) parseFlags() {
	flag.StringVar(&cfg.ListenAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.StoreInterval, "i", 4, "interval after which server will write data to disk")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json",
		"path to file in which server will store data")
	flag.BoolVar(&cfg.Restore, "r", true,
		"if false server will not restore data which it write before restart")
	flag.Parse()
}

func (cfg *config) parseEnv() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.ListenAddress = envRunAddr
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		storeInterval, err := strconv.Atoi(envStoreInterval)
		if err == nil {
			cfg.StoreInterval = storeInterval
		}
	}

	if value, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = value
	}
	fmt.Println(cfg.FileStoragePath)

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restore, err := strconv.ParseBool(envRestore)
		if err == nil {
			cfg.Restore = restore
		}
	}
}
