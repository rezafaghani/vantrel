package synthetic

import (
	"context"
	"testing"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/catalog"
	"github.com/rezafaghani/vantrel/libs/go/ingestion"
)

func TestRecordsAreDeterministic(t *testing.T) {
	cfg := Config{Seed: 42, BaseTime: time.Unix(1704067200, 0).UTC(), Steps: 2}
	a, err := Records(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Records(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 8 || len(b) != 8 {
		t.Fatalf("records=%d/%d", len(a), len(b))
	}
	for i := range a {
		if string(a[i].Payload) != string(b[i].Payload) || a[i].SourceSequence != b[i].SourceSequence {
			t.Fatalf("record %d differs\n%s\n%s", i, a[i].Payload, b[i].Payload)
		}
	}
}

func TestAdapterRunsThroughIngestionSDK(t *testing.T) {
	adapter, err := Adapter(Config{Seed: 7, Steps: 1})
	if err != nil {
		t.Fatal(err)
	}
	sink := &eventSink{}
	err = ingestion.Runtime{
		Adapter: adapter,
		Mode:    ingestion.ModeHistorical,
		Hooks: ingestion.Hooks{
			Parser:    Parser{},
			Publisher: sink,
		},
	}.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 4 {
		t.Fatalf("events=%d", len(sink.events))
	}
	wantTypes := map[string]bool{"trade": true, "quote": true, "forecast": true, "actual": true}
	for _, event := range sink.events {
		delete(wantTypes, event.Type)
		if event.SeriesID == "" || len(event.Payload) == 0 {
			t.Fatalf("bad event=%+v", event)
		}
	}
	if len(wantTypes) != 0 {
		t.Fatalf("missing event types=%v", wantTypes)
	}
}

func TestRegisterCatalog(t *testing.T) {
	store := catalog.NewMemoryStore()
	if err := RegisterCatalog(store); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{SeriesAppleTradePrice, SeriesAppleQuote, SeriesDK1WindForecast, SeriesDK1WindActual} {
		series, err := store.GetSeries(id)
		if err != nil {
			t.Fatal(err)
		}
		if !series.Active || series.ProviderID != ProviderID {
			t.Fatalf("series=%+v", series)
		}
	}
	relationships, err := store.GetRelationships(SeriesDK1WindForecast)
	if err != nil {
		t.Fatal(err)
	}
	if len(relationships) != 1 || relationships[0].Type != catalog.RelationshipForecastOf {
		t.Fatalf("relationships=%+v", relationships)
	}
	_, members, err := store.GetDataProduct(DataProductDK1PowerAndApple)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 4 {
		t.Fatalf("members=%d", len(members))
	}
}

type eventSink struct {
	events []ingestion.CanonicalEvent
}

func (s *eventSink) Publish(_ context.Context, event ingestion.CanonicalEvent) error {
	s.events = append(s.events, event)
	return nil
}
