package config

import (
	"flag"
	"os"
	"strconv"
)

type AgentConfig struct {
	Addr          string
	ReportSeconds int
	PollSeconds   int
}

func ParseAgentFlags() *AgentConfig {
	const (
		defaultAddr   = "http://localhost:8080"
		defaultReport = 10
		defaultPoll   = 2
	)

	addr := flag.String("a", defaultAddr, "server address")
	report := flag.Int("r", defaultReport, "report interval in seconds")
	poll := flag.Int("p", defaultPoll, "poll interval in seconds")
	flag.Parse()

	cfg := &AgentConfig{
		Addr:          *addr,
		ReportSeconds: *report,
		PollSeconds:   *poll,
	}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		if v, err := strconv.Atoi(envReport); err == nil {
			cfg.ReportSeconds = v
		}
	}
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if v, err := strconv.Atoi(envPoll); err == nil {
			cfg.PollSeconds = v
		}
	}

	return cfg
}
