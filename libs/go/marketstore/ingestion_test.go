package marketstore

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
	"github.com/rezafaghani/vantrel/libs/go/synthetic"
	"github.com/rezafaghani/vantrel/libs/go/validation"
)

func TestPublisherStoresValidatedSyntheticObservations(t *testing.T) {
	adapter, err := synthetic.Adapter(synthetic.Config{Seed: 3, Steps: 2})
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore()
	err = ingestion.Runtime{
		Adapter: adapter,
		Mode:    ingestion.ModeHistorical,
		Hooks: ingestion.Hooks{
			Parser:    synthetic.Parser{},
			Validator: validation.Engine{},
			Publisher: Publisher{Store: store},
		},
	}.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	actuals, err := store.QueryActual(Query{SeriesID: synthetic.SeriesDK1WindActual})
	if err != nil {
		t.Fatal(err)
	}
	forecasts, err := store.QueryForecast(Query{SeriesID: synthetic.SeriesDK1WindForecast})
	if err != nil {
		t.Fatal(err)
	}
	if len(actuals) != 2 || len(forecasts) != 2 {
		t.Fatalf("actuals=%d forecasts=%d", len(actuals), len(forecasts))
	}
}

func TestPublisherWritesSyntheticObservationsAsILP(t *testing.T) {
	records, err := synthetic.Records(synthetic.Config{Seed: 4, Steps: 1})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	writer := NewILPWriter(&buf)
	parser := synthetic.Parser{}
	publisher := Publisher{Store: writer}
	for _, record := range records {
		event, err := parser.Parse(context.Background(), record)
		if err != nil {
			t.Fatal(err)
		}
		if err := (validation.Engine{}).Validate(context.Background(), event); err != nil {
			t.Fatal(err)
		}
		if err := publisher.Publish(context.Background(), event); err != nil {
			t.Fatal(err)
		}
	}
	out := buf.String()
	if strings.Count(out, "\n") != 2 {
		t.Fatalf("ilp=%s", out)
	}
	if !strings.Contains(out, "market_actual_observations") || !strings.Contains(out, "market_forecast_observations") {
		t.Fatalf("ilp=%s", out)
	}
}
