package main

import (
	"testing"

	"honnef.co/go/tools/staticcheck"
)

func TestAnalyzers(t *testing.T) {
	checks := analyzers()
	names := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		if _, ok := names[check.Name]; ok {
			t.Fatalf("duplicate analyzer %q", check.Name)
		}
		names[check.Name] = struct{}{}
	}

	for _, check := range staticcheck.Analyzers {
		if _, ok := names[check.Analyzer.Name]; !ok {
			t.Errorf("staticcheck analyzer %q is not enabled", check.Analyzer.Name)
		}
	}

	for _, name := range []string{"printf", "S1000", "ST1005", "bodyclose", "nilerr", "noosexit"} {
		if _, ok := names[name]; !ok {
			t.Errorf("analyzer %q is not enabled", name)
		}
	}
}
