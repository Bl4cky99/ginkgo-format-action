// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Jason Giese (Bl4cky99)

package main

import (
	"fmt"
	"strings"
	"time"
)

type renderOptions struct {
	title              string
	showBreakdown      bool
	showFailureDetails bool
	maxFailureDetails  int // 0 = unlimited
}

var htmlReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

func render(reports []Report, c counts, dur time.Duration, failures []SpecReport, opts renderOptions) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## %s\n\n", opts.title)

	if c.total == 0 && len(failures) == 0 {
		b.WriteString("> [!NOTE]\n> No specs found in the report.\n\n")
		return b.String()
	}

	if c.failed > 0 {
		b.WriteString("> [!CAUTION]\n")
		b.WriteString("> **Ginkgo Test Suite Failed**\n")
	} else {
		b.WriteString("> [!NOTE]\n")
		b.WriteString("> **Ginkgo Test Suite Passed**\n")
	}
	fmt.Fprintf(&b, "> **%d** specs total │ Passed: **%d** │ Failed: **%d** │ Skipped: %d │ Pending: %d\n",
		c.total, c.passed, c.failed, c.skipped, c.pending)
	fmt.Fprintf(&b, "> *Total Duration: %s*\n\n", dur.Round(time.Millisecond))

	if opts.showFailureDetails && len(failures) > 0 {
		renderFailureSection(&b, failures, opts.maxFailureDetails)
	}

	if opts.showBreakdown && len(reports) > 0 {
		renderBreakdownTable(&b, reports)
	}

	return b.String()
}

func renderFailureSection(b *strings.Builder, failures []SpecReport, max int) {
	shown := failures
	if max > 0 && len(failures) > max {
		shown = failures[:max]
	}

	b.WriteString("### ❌ Failure Details\n\n")
	for _, s := range shown {
		renderFailureBlock(b, s)
	}

	if len(shown) < len(failures) {
		fmt.Fprintf(b, "> [!NOTE]\n> Showing %d of %d failures. Increase `max-failure-details` or set it to `0` to see all.\n\n",
			len(shown), len(failures))
	}

	b.WriteString("---\n\n")
}

func renderBreakdownTable(b *strings.Builder, reports []Report) {
	b.WriteString("### Breakdown\n\n")
	b.WriteString("| Result | Suite | Specs | Passed | Failed | Skipped | Pending | Duration |\n")
	b.WriteString("| :---: | :--- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, r := range reports {
		var sc counts
		for _, s := range r.SpecReports {
			if s.LeafNodeType == "It" {
				sc.add(s.State)
			}
		}
		icon := "✔"
		if !r.SuiteSucceeded {
			icon = "❌"
		}
		name := firstNonEmpty(r.SuiteDescription, r.SuitePath)
		fmt.Fprintf(b, "| %s | %s | %d | %d | %d | %d | %d | %s |\n",
			icon, escapePipe(name), sc.total, sc.passed, sc.failed, sc.skipped, sc.pending,
			r.RunTime.Round(time.Millisecond))
	}
	b.WriteString("\n")
}

func renderFailureBlock(b *strings.Builder, s SpecReport) {
	msg := strings.TrimSpace(s.Failure.Message)
	if msg == "" {
		msg = "(no failure message)"
	}

	contextName := firstNonEmpty(
		sliceFirst(s.ContainerHierarchyTexts),
		s.LeafNodeText,
		fmt.Sprintf("[%s]", s.LeafNodeType),
	)

	locationStr := ""
	if s.Failure.Location.FileName != "" {
		locationStr = fmt.Sprintf(" — *%s:%d*", s.Failure.Location.FileName, s.Failure.Location.LineNumber)
	}

	fmt.Fprintf(b, "<details>\n<summary><code>%s</code> <b>%s</b>%s</summary>\n<br>\n\n",
		getFailureType(s.State), escapeHTML(contextName), locationStr)

	b.WriteString("| Property | Details |\n")
	b.WriteString("| :--- | :--- |\n")
	fmt.Fprintf(b, "| **Spec** | ` %s ` |\n", escapeHTML(s.FullText()))
	fmt.Fprintf(b, "| **State** | ` %s ` |\n", s.State)

	if !strings.Contains(msg, "\n") {
		fmt.Fprintf(b, "| **Error** | ` %s ` |\n\n", fenceSafe(msg))
	} else {
		fmt.Fprintf(b, "| **Error** | <pre>%s</pre> |\n\n",
			strings.ReplaceAll(fenceSafe(msg), "\n", "<br>"))
	}

	b.WriteString("</details>\n\n")
}

func getFailureType(state string) string {
	switch state {
	case "panicked":
		return "CRASH"
	case "timedout", "interrupted":
		return "TIMEOUT"
	case "aborted":
		return "ABORTED"
	default:
		return "FAIL"
	}
}

func escapeHTML(s string) string {
	return htmlReplacer.Replace(s)
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

func fenceSafe(s string) string {
	return strings.ReplaceAll(s, "```", "`​``")
}

func sliceFirst(ss []string) string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}
