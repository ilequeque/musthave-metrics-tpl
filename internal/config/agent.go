package config

import "flag"

type AgentConfig struct {
	Addr          string
	ReportSeconds int
	PollSeconds   int
}

func ParseAgentFlags() AgentConfig {
	addr := flag.String("a", "http://localhost:8080", "server address")
	report := flag.Int("r", 10, "report interval in seconds")
	poll := flag.Int("p", 2, "poll interval in seconds")
	flag.Parse()
	return AgentConfig{
		Addr:          *addr,
		ReportSeconds: *report,
		PollSeconds:   *poll,
	}
}
