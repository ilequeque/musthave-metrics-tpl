package config

import (
	"flag"
	"os"
	"strconv"
)

type ServerConfig struct {
	Addr          string
	StoreInterval int
	FileStorage   string
	Restore       bool
}

func ParseServerFlags() ServerConfig {
	addr := flag.String("a", "localhost:8080", "address for HTTP server")
	store := flag.Int("i", 300, "store interval in seconds (0 = sync write)")
	file := flag.String("f", "/tmp/metrics-db.json", "path to file for metrics storage")
	restore := flag.Bool("r", true, "restore metrics from file on startup")
	flag.Parse()
	
	cfg := ServerConfig{
		Addr:          *addr,
		StoreInterval: *store,
		FileStorage:   *file,
		Restore:       *restore,
	}

	if v := os.Getenv("ADDRESS"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("STORE_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.StoreInterval = n
		}
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.FileStorage = v
	}
	if v := os.Getenv("RESTORE"); v != "" {
		cfg.Restore = v == "true"
	}

	return cfg
}
