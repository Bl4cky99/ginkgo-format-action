// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Jason Giese (Bl4cky99)

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	ReportPath         string `envconfig:"REPORT_PATH"          default:"report.json"`
	FailOnFailures     bool   `envconfig:"FAIL_ON_FAILURES"`
	Title              string `envconfig:"TITLE"                default:"Ginkgo Test Results"`
	ShowBreakdown      bool   `envconfig:"SHOW_BREAKDOWN"       default:"true"`
	ShowFailureDetails bool   `envconfig:"SHOW_FAILURE_DETAILS" default:"true"`
	MaxFailureDetails  int    `envconfig:"MAX_FAILURE_DETAILS"`
}

func loadConfig(args []string) (config, error) {
	normalizeInputEnv()
	var cfg config
	if err := envconfig.Process("INPUT", &cfg); err != nil {
		return config{}, fmt.Errorf("config: %w", err)
	}
	if cfg.MaxFailureDetails < 0 {
		return config{}, fmt.Errorf("config: INPUT_MAX_FAILURE_DETAILS must be >= 0, got %d", cfg.MaxFailureDetails)
	}
	if len(args) > 0 {
		if v := strings.TrimSpace(args[0]); v != "" && !strings.HasPrefix(v, "-") {
			cfg.ReportPath = v
		}
	}
	return cfg, nil
}

func normalizeInputEnv() {
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "INPUT_") || !strings.Contains(kv, "-") {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		normalized := strings.ReplaceAll(parts[0], "-", "_")
		if _, exists := os.LookupEnv(normalized); !exists {
			os.Setenv(normalized, parts[1])
		}
	}
}

func (c config) renderOpts() renderOptions {
	return renderOptions{
		title:              c.Title,
		showBreakdown:      c.ShowBreakdown,
		showFailureDetails: c.ShowFailureDetails,
		maxFailureDetails:  c.MaxFailureDetails,
	}
}
