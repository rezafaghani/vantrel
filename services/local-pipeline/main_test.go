package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunPrintsPipelineSummary(t *testing.T) {
	var out bytes.Buffer
	err := run(context.Background(), 2, time.Unix(1704067200, 0).UTC(), &out)
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"vantrel local pipeline",
		"steps=2",
		"market_ilp_lines=8",
		"first_ilp_line=market_trade_observations",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRunRejectsBadSteps(t *testing.T) {
	if err := run(context.Background(), 0, time.Unix(1704067200, 0).UTC(), &bytes.Buffer{}); err == nil {
		t.Fatal("expected error")
	}
}
