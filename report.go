// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Jason Giese (Bl4cky99)

package main

import (
	"fmt"
	"strings"
	"time"
)

type Report struct {
	SuiteDescription string        `json:"SuiteDescription"`
	SuitePath        string        `json:"SuitePath"`
	SuiteSucceeded   bool          `json:"SuiteSucceeded"`
	RunTime          time.Duration `json:"RunTime"`
	SpecReports      []SpecReport  `json:"SpecReports"`
}

type SpecReport struct {
	ContainerHierarchyTexts []string      `json:"ContainerHierarchyTexts"`
	LeafNodeText            string        `json:"LeafNodeText"`
	LeafNodeType            string        `json:"LeafNodeType"`
	State                   string        `json:"State"`
	RunTime                 time.Duration `json:"RunTime"`
	Failure                 Failure       `json:"Failure"`
}

type Failure struct {
	Message  string       `json:"Message"`
	Location CodeLocation `json:"Location"`
}

type CodeLocation struct {
	FileName   string `json:"FileName"`
	LineNumber int    `json:"LineNumber"`
}

type counts struct {
	total, passed, failed, skipped, pending int
}

func (c *counts) add(state string) {
	c.total++
	switch state {
	case "passed":
		c.passed++
	case "skipped":
		c.skipped++
	case "pending":
		c.pending++
	default:
		if isFailureState(state) {
			c.failed++
		}
	}
}

func (s SpecReport) FullText() string {
	parts := make([]string, 0, len(s.ContainerHierarchyTexts)+1)
	parts = append(parts, s.ContainerHierarchyTexts...)
	if s.LeafNodeText != "" {
		parts = append(parts, s.LeafNodeText)
	}
	if len(parts) == 0 {
		return fmt.Sprintf("[%s]", s.LeafNodeType)
	}
	return strings.Join(parts, " › ")
}

func isFailureState(state string) bool {
	switch state {
	case "failed", "panicked", "timedout", "aborted", "interrupted":
		return true
	}
	return false
}
