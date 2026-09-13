package marketstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
)

type Publisher struct {
	Store ObservationWriter
}

func (p Publisher) Publish(_ context.Context, event ingestion.CanonicalEvent) error {
	if p.Store == nil {
		return fmt.Errorf("%w: store is required", ErrInvalid)
	}
	if event.Type != "actual" && event.Type != "forecast" {
		return nil
	}
	var payload observationPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}
	lineage := Lineage{
		RawRecordID:      event.Metadata["raw_record_id"],
		CanonicalEventID: event.ID,
		IngestionID:      event.Metadata["ingestion_id"],
	}
	if event.Type == "actual" {
		return p.Store.WriteActual(ActualObservation{
			SeriesID:  first(payload.SeriesID, event.SeriesID),
			EventTime: payload.EventTime,
			Value:     payload.Value,
			Unit:      payload.Unit,
			Quality:   Quality{State: "VALID"},
			Lineage:   lineage,
		})
	}
	return p.Store.WriteForecast(ForecastObservation{
		SeriesID:      first(payload.SeriesID, event.SeriesID),
		ForecastRunID: payload.ForecastRunID,
		IssuedAt:      payload.IssuedAt,
		TargetTime:    payload.TargetTime,
		Horizon:       time.Duration(payload.HorizonMinutes) * time.Minute,
		Value:         payload.Value,
		Unit:          payload.Unit,
		ModelVersion:  payload.ModelVersion,
		Quality:       Quality{State: "VALID"},
		Lineage:       lineage,
	})
}

type observationPayload struct {
	SeriesID       string    `json:"series_id"`
	EventTime      time.Time `json:"event_time"`
	Value          float64   `json:"value"`
	Unit           string    `json:"unit"`
	ForecastRunID  string    `json:"forecast_run_id"`
	IssuedAt       time.Time `json:"issued_at"`
	TargetTime     time.Time `json:"target_time"`
	HorizonMinutes int       `json:"horizon_minutes"`
	ModelVersion   string    `json:"model_version"`
}

func first(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
