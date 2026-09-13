package marketstore

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestILPWriterWritesActualLine(t *testing.T) {
	var buf bytes.Buffer
	err := NewILPWriter(&buf).WriteActual(ActualObservation{
		SeriesID:  "SYN_EQ_AAPL_TRADE_PRICE",
		EventTime: time.Unix(1, 2).UTC(),
		Value:     185.12,
		Unit:      "USD",
		Quality:   Quality{State: "VALID", Score: 0.99, Flags: []string{"synthetic"}},
		Revision:  2,
		Lineage:   Lineage{RawRecordID: "raw-1", CanonicalEventID: "event-1", IngestionID: "ing-1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "market_actual_observations,series_id=SYN_EQ_AAPL_TRADE_PRICE,unit=USD,quality_state=VALID,raw_record_id=raw-1,canonical_event_id=event-1,ingestion_id=ing-1 value=185.12,quality_score=0.99,revision=2i,quality_flags=\"synthetic\" 1000000002\n"
	if buf.String() != want {
		t.Fatalf("line=\n%s", buf.String())
	}
}

func TestILPWriterWritesForecastLine(t *testing.T) {
	var buf bytes.Buffer
	t0 := time.Unix(1704067200, 0).UTC()
	err := NewILPWriter(&buf).WriteForecast(ForecastObservation{
		SeriesID:      "SYN_DK1_WIND_FORECAST",
		ForecastRunID: "run 1",
		IssuedAt:      t0,
		TargetTime:    t0.Add(time.Hour),
		Horizon:       time.Hour,
		Value:         620.5,
		Unit:          "MW",
		ModelVersion:  "model,a",
	})
	if err != nil {
		t.Fatal(err)
	}
	line := buf.String()
	if !strings.HasPrefix(line, `market_forecast_observations,series_id=SYN_DK1_WIND_FORECAST,forecast_run_id=run\ 1,unit=MW,model_version=model\,a issued_at=1704067200000000t,horizon_seconds=3600i,value=620.5`) {
		t.Fatalf("line=%s", line)
	}
	if !strings.HasSuffix(line, "1704070800000000000\n") {
		t.Fatalf("line=%s", line)
	}
}

func TestILPWriterWritesTradeAndQuoteLines(t *testing.T) {
	var buf bytes.Buffer
	writer := NewILPWriter(&buf)
	t0 := time.Unix(1704067200, 0).UTC()
	if err := writer.WriteTrade(TradeObservation{SeriesID: "SYN_EQ_AAPL_TRADE_PRICE", TradeTime: t0, Price: 185.12, Quantity: 100, Currency: "USD", Instrument: "AAPL"}); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteQuote(QuoteObservation{SeriesID: "SYN_EQ_AAPL_QUOTE", QuoteTime: t0, Bid: 185.1, Ask: 185.14, Quantity: 500, Currency: "USD", Instrument: "AAPL"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "market_trade_observations,series_id=SYN_EQ_AAPL_TRADE_PRICE,instrument=AAPL,currency=USD price=185.12,quantity=100") {
		t.Fatalf("ilp=%s", out)
	}
	if !strings.Contains(out, "market_quote_observations,series_id=SYN_EQ_AAPL_QUOTE,instrument=AAPL,currency=USD bid=185.1,ask=185.14,quantity=500") {
		t.Fatalf("ilp=%s", out)
	}
}

func TestILPWriterRejectsInvalidObservation(t *testing.T) {
	var buf bytes.Buffer
	err := NewILPWriter(&buf).WriteActual(ActualObservation{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
