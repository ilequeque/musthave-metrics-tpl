package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Addr string
}

func ParseServerFlags() *ServerConfig {
	const defaultAddr = "localhost:8080"

	addr := flag.String("a", defaultAddr, "address for HTTP server")
	flag.Parse()

	cfg := &ServerConfig{Addr: *addr}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}

	return cfg
}
