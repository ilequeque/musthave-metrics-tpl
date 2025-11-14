package config

import "flag"

type ServerConfig struct {
	Addr string
}

func ParseServerFlags() ServerConfig {
	addr := flag.String("a", "localhost:8080", "address for HTTP server")
	flag.Parse()
	return ServerConfig{Addr: *addr}
}
