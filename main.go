// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Jason Giese (Bl4cky99)

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	cfg, err := loadConfig(os.Args[1:])
	if err != nil {
		fatalf("%v", err)
	}

	path := cfg.ReportPath
	if !filepath.IsAbs(path) {
		if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
			path = filepath.Join(ws, path)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("::warning::Ginkgo report not found at %s — skipping summary.\n", path)
			writeSummary(fmt.Sprintf(
				"## %s\n\n> No reports found at `%s`. Did the Ginkgo step run with `--json-report`?\n\n",
				cfg.Title, path,
			))
			return
		}
		fatalf("could not read report %q: %v", path, err)
	}

	var reports []Report
	if err := json.Unmarshal(data, &reports); err != nil {
		fatalf("could not parse report %q: %v", path, err)
	}

	overall, totalDuration, failures := collectStats(reports)

	writeSummary(render(reports, overall, totalDuration, failures, cfg.renderOpts()))

	writeOutputs(map[string]string{
		"total":     strconv.Itoa(overall.total),
		"passed":    strconv.Itoa(overall.passed),
		"failed":    strconv.Itoa(overall.failed),
		"skipped":   strconv.Itoa(overall.skipped),
		"pending":   strconv.Itoa(overall.pending),
		"succeeded": strconv.FormatBool(overall.failed == 0),
	})

	if overall.failed > 0 && cfg.FailOnFailures {
		os.Exit(1)
	}
}

func collectStats(reports []Report) (overall counts, totalDuration time.Duration, failures []SpecReport) {
	for _, r := range reports {
		totalDuration += r.RunTime
		for _, s := range r.SpecReports {
			if s.LeafNodeType == "It" {
				overall.add(s.State)
			}
			if isFailureState(s.State) {
				failures = append(failures, s)
			}
		}
	}
	return
}

func writeSummary(md string) {
	if p := os.Getenv("GITHUB_STEP_SUMMARY"); p != "" {
		appendToFile(p, func(f *os.File) {
			if _, err := f.WriteString(md); err != nil {
				fatalf("could not write step summary: %v", err)
			}
		})
		return
	}
	fmt.Print(md)
}

func writeOutputs(outputs map[string]string) {
	p := os.Getenv("GITHUB_OUTPUT")
	if p == "" {
		return
	}
	appendToFile(p, func(f *os.File) {
		for k, v := range outputs {
			fmt.Fprintf(f, "%s=%s\n", k, v)
		}
	})
}

func appendToFile(path string, fn func(*os.File)) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		fatalf("could not open file %s: %v", path, err)
	}
	defer f.Close()
	fn(f)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ginkgo-summary: "+format+"\n", args...)
	os.Exit(1)
}
