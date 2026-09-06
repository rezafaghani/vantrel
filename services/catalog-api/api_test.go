package catalogapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rezafaghani/vantrel/libs/go/catalog"
)

func TestCatalogAPIFlow(t *testing.T) {
	store := catalog.NewMemoryStore()
	handler := New(store, nil).Handler()

	post(t, handler, "/v1/providers", catalog.Provider{ID: "energinet", DisplayName: "Energinet"}, http.StatusOK)
	post(t, handler, "/v1/sources", catalog.Source{ID: "energinet-api", ProviderID: "energinet", Name: "Energinet API", Protocol: "REST"}, http.StatusOK)
	post(t, handler, "/v1/series", seriesWrite{Series: baseSeries("dk1_wind_actual", "DK1 wind actual"), Reason: "initial"}, http.StatusOK)
	post(t, handler, "/v1/series", seriesWrite{Series: baseSeries("dk1_wind_forecast", "DK1 wind forecast"), Reason: "initial"}, http.StatusOK)
	post(t, handler, "/v1/relationships", catalog.SeriesRelationship{FromSeriesID: "dk1_wind_forecast", ToSeriesID: "dk1_wind_actual", Type: catalog.RelationshipForecastOf}, http.StatusOK)
	post(t, handler, "/v1/data-products", catalog.DataProduct{ID: "DK1_SHORT_TERM_POWER_FUNDAMENTALS", DisplayName: "DK1 fundamentals", Purposes: []catalog.Purpose{catalog.PurposePreTrade}}, http.StatusOK)
	post(t, handler, "/v1/data-products/DK1_SHORT_TERM_POWER_FUNDAMENTALS/members", catalog.DataProductMember{SeriesID: "dk1_wind_actual", Role: "actual_generation", Required: true}, http.StatusOK)

	updated := baseSeries("ignored-by-route", "DK1 wind actual")
	updated.Description = "Updated through API"
	put(t, handler, "/v1/series/dk1_wind_actual", seriesWrite{Series: updated, Reason: "description"}, http.StatusOK)

	res := get(t, handler, "/v1/series?q=forecast", http.StatusOK)
	var series []catalog.Series
	decode(t, res, &series)
	if len(series) != 1 || series[0].ID != "dk1_wind_forecast" {
		t.Fatalf("series=%+v", series)
	}

	res = get(t, handler, "/v1/series/dk1_wind_actual/relationships", http.StatusOK)
	var relationships []catalog.SeriesRelationship
	decode(t, res, &relationships)
	if len(relationships) != 1 || relationships[0].Type != catalog.RelationshipForecastOf {
		t.Fatalf("relationships=%+v", relationships)
	}

	res = get(t, handler, "/v1/data-products/DK1_SHORT_TERM_POWER_FUNDAMENTALS", http.StatusOK)
	var product struct {
		Members []catalog.DataProductMember `json:"members"`
	}
	decode(t, res, &product)
	if len(product.Members) != 1 || product.Members[0].SeriesID != "dk1_wind_actual" {
		t.Fatalf("product=%+v", product)
	}

	res = get(t, handler, "/v1/series/dk1_wind_actual/versions", http.StatusOK)
	var versions []catalog.SeriesVersion
	decode(t, res, &versions)
	if len(versions) != 2 || versions[1].Snapshot.Description != "Updated through API" {
		t.Fatalf("versions=%+v", versions)
	}
}

func TestCatalogAPIAuthHook(t *testing.T) {
	handler := New(catalog.NewMemoryStore(), func(*http.Request, string) bool { return false }).Handler()
	post(t, handler, "/v1/providers", catalog.Provider{ID: "x", DisplayName: "x"}, http.StatusForbidden)
}

func TestCatalogAPIReturnsValidationErrors(t *testing.T) {
	handler := New(catalog.NewMemoryStore(), nil).Handler()
	post(t, handler, "/v1/providers", catalog.Provider{ID: "missing-name"}, http.StatusBadRequest)
}

func post(t *testing.T, h http.Handler, path string, body any, want int) []byte {
	t.Helper()
	return requestJSON(t, h, http.MethodPost, path, body, want)
}

func put(t *testing.T, h http.Handler, path string, body any, want int) []byte {
	t.Helper()
	return requestJSON(t, h, http.MethodPut, path, body, want)
}

func requestJSON(t *testing.T, h http.Handler, method, path string, body any, want int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return do(t, h, req, want)
}

func get(t *testing.T, h http.Handler, path string, want int) []byte {
	t.Helper()
	return do(t, h, httptest.NewRequest(http.MethodGet, path, nil), want)
}

func do(t *testing.T, h http.Handler, req *http.Request, want int) []byte {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s status=%d body=%s", req.Method, req.URL.Path, rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

func decode(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

func baseSeries(id, name string) catalog.Series {
	return catalog.Series{
		ID:            id,
		CanonicalName: name,
		DisplayName:   name,
		AssetClass:    "POWER",
		SeriesType:    "ACTUAL",
		ProviderID:    "energinet",
		SourceID:      "energinet-api",
		Market:        "DK1",
		BiddingZone:   "DK1",
		Resolution:    "PT15M",
		Timezone:      "Europe/Copenhagen",
		Unit:          "MW",
		ValueType:     "DECIMAL",
		Purposes:      []catalog.Purpose{catalog.PurposePreTrade, catalog.PurposeRealtime},
		Active:        true,
	}
}
