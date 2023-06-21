package config

import (
	"flag"
	"os"
)

type config struct {
	ListenAddress string
}

func New() config {
	cfg := config{}
	cfg.parseFlags()
	cfg.parseEnv()
	return cfg
}
func (cfg *config) parseFlags() {
	flag.StringVar(&cfg.ListenAddress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}

func (cfg *config) parseEnv() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.ListenAddress = envRunAddr
	}
}
