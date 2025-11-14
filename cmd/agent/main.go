package main

import (
	"log"
	"time"

	"github.com/ilequeque/musthave-metrics-tpl/internal/agent"
	"github.com/ilequeque/musthave-metrics-tpl/internal/config"
)

func main() {
	cfg := config.ParseAgentFlags()

	r := agent.NewRunner(
		cfg.Addr,
		time.Duration(cfg.PollSeconds)*time.Second,
		time.Duration(cfg.ReportSeconds)*time.Second,
	)
	defer r.Close()

	log.Printf("agent started: addr=%s poll=%ds report=%ds",
		cfg.Addr, cfg.PollSeconds, cfg.ReportSeconds)

	r.Run()
}
