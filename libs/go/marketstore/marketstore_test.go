package marketstore

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestMemoryStoreKeepsActualAndForecastSeparate(t *testing.T) {
	store := NewMemoryStore()
	t0 := time.Unix(1704067200, 0).UTC()
	if err := store.WriteActual(ActualObservation{SeriesID: "actual", EventTime: t0, Value: 10, Unit: "MW"}); err != nil {
		t.Fatal(err)
	}
	if err := store.WriteForecast(ForecastObservation{SeriesID: "forecast", ForecastRunID: "run-1", IssuedAt: t0, TargetTime: t0.Add(time.Hour), Horizon: time.Hour, Value: 12, Unit: "MW"}); err != nil {
		t.Fatal(err)
	}

	actuals, err := store.QueryActual(Query{SeriesID: "actual"})
	if err != nil {
		t.Fatal(err)
	}
	forecasts, err := store.QueryForecast(Query{SeriesID: "forecast"})
	if err != nil {
		t.Fatal(err)
	}
	if len(actuals) != 1 || len(forecasts) != 1 {
		t.Fatalf("actuals=%d forecasts=%d", len(actuals), len(forecasts))
	}
}

func TestQueryFiltersAndSortsByTime(t *testing.T) {
	store := NewMemoryStore()
	t0 := time.Unix(1704067200, 0).UTC()
	for _, obs := range []ActualObservation{
		{SeriesID: "s", EventTime: t0.Add(2 * time.Hour), Value: 2, Unit: "MW"},
		{SeriesID: "other", EventTime: t0.Add(time.Hour), Value: 99, Unit: "MW"},
		{SeriesID: "s", EventTime: t0.Add(time.Hour), Value: 1, Unit: "MW"},
	} {
		if err := store.WriteActual(obs); err != nil {
			t.Fatal(err)
		}
	}
	got, err := store.QueryActual(Query{SeriesID: "s", From: t0.Add(30 * time.Minute), To: t0.Add(2 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Value != 1 || got[1].Value != 2 {
		t.Fatalf("observations=%+v", got)
	}
}

func TestValidateForecastRejectsMissingRun(t *testing.T) {
	err := ValidateForecast(ForecastObservation{SeriesID: "forecast", IssuedAt: time.Now(), TargetTime: time.Now(), Unit: "MW"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestValidationRejectsBadValuesAndRanges(t *testing.T) {
	if err := ValidateActual(ActualObservation{SeriesID: "actual", EventTime: time.Now(), Value: math.NaN(), Unit: "MW"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
	if err := ValidateQuery(Query{SeriesID: "actual", From: time.Unix(2, 0), To: time.Unix(1, 0)}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}
