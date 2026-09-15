package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestQuestDBInitWriteAndStatus(t *testing.T) {
	roundTrip := func(r *http.Request) *http.Response {
		switch r.URL.Path {
		case "/exec":
			return response(`{"columns":[{"name":"count"}],"dataset":[[2]]}`)
		case "/write":
			return response(`{}`)
		default:
			return &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found", Body: io.NopCloser(strings.NewReader(""))}
		}
	}

	db := questDB{base: "http://questdb.test", client: &http.Client{Transport: roundTripper(roundTrip)}}
	if err := db.init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := db.writeILP(context.Background(), "market_trade_observations price=1 1\n"); err != nil {
		t.Fatal(err)
	}
	status := db.status(context.Background())
	if status["questdb_ok"] != true || status["trades"] != 2 {
		t.Fatalf("status=%v", status)
	}
	latest := db.latest(context.Background())
	if len(latest) != 4 {
		t.Fatalf("latest=%v", latest)
	}
}

func TestRowsMapsColumnsToValues(t *testing.T) {
	got := rows([]columnInfo{{Name: "series_id"}, {Name: "price"}}, [][]any{{"S", 12.3}})
	if len(got) != 1 || got[0]["series_id"] != "S" || got[0]["price"] != 12.3 {
		t.Fatalf("rows=%v", got)
	}
}

func TestRunEndpointRequiresPost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/run", nil)
	rec := httptest.NewRecorder()
	newMux(questDB{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d", rec.Code)
	}
}

type roundTripper func(*http.Request) *http.Response

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r), nil
}

func response(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(body))}
}
