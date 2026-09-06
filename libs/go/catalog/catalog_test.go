package catalog

import (
	"errors"
	"testing"
)

func TestCatalogStoresVersionedSeriesRelationshipsAndProducts(t *testing.T) {
	store := NewMemoryStore()
	must(t, store.SaveProvider(Provider{ID: "energinet", DisplayName: "Energinet", License: "open-data"}))
	must(t, store.SaveSource(Source{ID: "energinet-api", ProviderID: "energinet", Name: "Energinet API", Protocol: "REST"}))

	actual := baseSeries("dk1_wind_actual", "DK1 wind actual")
	forecast := baseSeries("dk1_wind_forecast", "DK1 wind forecast")
	forecast.SeriesType = "FORECAST"

	version, err := store.SaveSeries(actual, "test", "initial")
	if err != nil {
		t.Fatal(err)
	}
	if version.Version != 1 {
		t.Fatalf("version=%d, want 1", version.Version)
	}
	must(t, mustVersion(store.SaveSeries(forecast, "test", "initial")))
	must(t, store.AddRelationship(SeriesRelationship{
		FromSeriesID: forecast.ID,
		ToSeriesID:   actual.ID,
		Type:         RelationshipForecastOf,
	}))
	must(t, store.SaveDataProduct(DataProduct{
		ID:          "DK1_SHORT_TERM_POWER_FUNDAMENTALS",
		DisplayName: "DK1 short-term power fundamentals",
		Purposes:    []Purpose{PurposePreTrade, PurposeTraining},
		Active:      true,
	}))
	must(t, store.AddDataProductMember(DataProductMember{
		DataProductID: "DK1_SHORT_TERM_POWER_FUNDAMENTALS",
		SeriesID:      actual.ID,
		Role:          "actual_generation",
		Required:      true,
	}))

	got, err := store.GetSeries(forecast.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.SeriesType != "FORECAST" || got.Purposes[0] != PurposePreTrade {
		t.Fatalf("series=%+v", got)
	}
	rels, err := store.GetRelationships(actual.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rels) != 1 || rels[0].Type != RelationshipForecastOf {
		t.Fatalf("relationships=%+v", rels)
	}
	product, members, err := store.GetDataProduct("DK1_SHORT_TERM_POWER_FUNDAMENTALS")
	if err != nil {
		t.Fatal(err)
	}
	if product.ID == "" || len(members) != 1 || members[0].SeriesID != actual.ID {
		t.Fatalf("product=%+v members=%+v", product, members)
	}
}

func TestSeriesVersionAuditTrail(t *testing.T) {
	store := NewMemoryStore()
	must(t, store.SaveProvider(Provider{ID: "nasdaq", DisplayName: "Nasdaq"}))
	must(t, store.SaveSource(Source{ID: "nasdaq-basic", ProviderID: "nasdaq", Name: "Nasdaq Basic", Protocol: "WebSocket"}))
	series := baseSeries("aapl_trade_price", "Apple trade price")
	series.ProviderID = "nasdaq"
	series.SourceID = "nasdaq-basic"
	series.AssetClass = "EQUITY"
	series.Instrument = "AAPL"
	must(t, mustVersion(store.SaveSeries(series, "test", "initial")))

	series.Description = "Apple trade price observations"
	must(t, mustVersion(store.SaveSeries(series, "test", "description")))

	versions, err := store.GetSeriesVersions(series.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0].Version != 1 || versions[1].Version != 2 {
		t.Fatalf("versions=%+v", versions)
	}
	if versions[0].Snapshot.Description == versions[1].Snapshot.Description {
		t.Fatalf("snapshots were not versioned: %+v", versions)
	}
}

func TestSearchSeries(t *testing.T) {
	store := NewMemoryStore()
	must(t, store.SaveProvider(Provider{ID: "energinet", DisplayName: "Energinet"}))
	must(t, store.SaveSource(Source{ID: "energinet-api", ProviderID: "energinet", Name: "Energinet API", Protocol: "REST"}))
	series := baseSeries("dk1_wind_actual", "DK1 wind actual")
	series.Tags = []string{"wind", "generation"}
	must(t, mustVersion(store.SaveSeries(series, "test", "initial")))

	got, err := store.SearchSeries("generation")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != series.ID {
		t.Fatalf("series=%+v", got)
	}
}

func TestValidationRejectsBadReferences(t *testing.T) {
	store := NewMemoryStore()
	if err := store.SaveSource(Source{ID: "missing-provider-source", ProviderID: "missing", Name: "x", Protocol: "REST"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
	if _, err := store.SaveSeries(baseSeries("", "missing id"), "test", "bad"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func baseSeries(id, name string) Series {
	return Series{
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
		Purposes:      []Purpose{PurposePreTrade, PurposeRealtime, PurposeTraining},
		Active:        true,
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustVersion(_ SeriesVersion, err error) error {
	return err
}
