package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type questDB struct {
	base       string
	schemaPath string
	client     *http.Client
}

func (q questDB) init(ctx context.Context) error {
	path := q.schemaPath
	if path == "" {
		path = "../../contracts/questdb/market_observations.sql"
	}
	sql, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, stmt := range strings.Split(string(sql), ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if err := q.exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (q questDB) writeILP(ctx context.Context, lines string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.base+"/write?precision=n", strings.NewReader(lines))
	if err != nil {
		return err
	}
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("questdb write failed: %s", resp.Status)
	}
	return nil
}

func (q questDB) status(ctx context.Context) map[string]any {
	out := map[string]any{"questdb_ok": true}
	for name, table := range observationTables {
		count, err := q.count(ctx, table)
		if err != nil {
			out["questdb_ok"] = false
			out[name] = 0
			continue
		}
		out[name] = count
	}
	return out
}

func (q questDB) latest(ctx context.Context) map[string]any {
	out := map[string]any{}
	for name, table := range observationTables {
		var body struct {
			Columns []columnInfo `json:"columns"`
			Dataset [][]any      `json:"dataset"`
		}
		if err := q.query(ctx, "select * from "+table+" limit -5", &body); err != nil {
			out[name] = []any{}
			continue
		}
		out[name] = rows(body.Columns, body.Dataset)
	}
	return out
}

type columnInfo struct {
	Name string `json:"name"`
}

func rows(columns []columnInfo, dataset [][]any) []map[string]any {
	out := make([]map[string]any, 0, len(dataset))
	for _, values := range dataset {
		row := map[string]any{}
		for i, column := range columns {
			if i < len(values) {
				row[column.Name] = values[i]
			}
		}
		out = append(out, row)
	}
	return out
}

func (q questDB) count(ctx context.Context, table string) (int, error) {
	var body struct {
		Dataset [][]any `json:"dataset"`
	}
	if err := q.query(ctx, "select count() from "+table, &body); err != nil {
		return 0, err
	}
	if len(body.Dataset) == 0 || len(body.Dataset[0]) == 0 {
		return 0, nil
	}
	switch v := body.Dataset[0][0].(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, nil
	}
}

var observationTables = map[string]string{
	"actuals":   "market_actual_observations",
	"forecasts": "market_forecast_observations",
	"trades":    "market_trade_observations",
	"quotes":    "market_quote_observations",
}

func (q questDB) exec(ctx context.Context, sql string) error {
	return q.query(ctx, sql, &struct{}{})
}

func (q questDB) query(ctx context.Context, sql string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, q.base+"/exec?query="+url.QueryEscape(sql), nil)
	if err != nil {
		return err
	}
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("questdb query failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buf.Bytes())
}
