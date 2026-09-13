package synthetic

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
)

const (
	ProviderID = "synthetic"
	SourceID   = "synthetic-market-source"

	SeriesAppleTradePrice = "SYN_EQ_AAPL_TRADE_PRICE"
	SeriesAppleQuote      = "SYN_EQ_AAPL_QUOTE"
	SeriesDK1WindForecast = "SYN_DK1_WIND_FORECAST"
	SeriesDK1WindActual   = "SYN_DK1_WIND_ACTUAL"
)

type Config struct {
	Seed     int64
	BaseTime time.Time
	Steps    int
}

type Payload struct {
	Type           string    `json:"type"`
	Sequence       string    `json:"sequence"`
	SeriesID       string    `json:"series_id"`
	EventTime      time.Time `json:"event_time"`
	Symbol         string    `json:"symbol,omitempty"`
	BiddingZone    string    `json:"bidding_zone,omitempty"`
	Price          float64   `json:"price,omitempty"`
	Bid            float64   `json:"bid,omitempty"`
	Ask            float64   `json:"ask,omitempty"`
	Quantity       float64   `json:"quantity,omitempty"`
	Value          float64   `json:"value,omitempty"`
	Unit           string    `json:"unit,omitempty"`
	ForecastRunID  string    `json:"forecast_run_id,omitempty"`
	IssuedAt       time.Time `json:"issued_at,omitempty"`
	TargetTime     time.Time `json:"target_time,omitempty"`
	HorizonMinutes int       `json:"horizon_minutes,omitempty"`
}

func Records(cfg Config) ([]ingestion.Record, error) {
	cfg = defaults(cfg)
	rng := rand.New(rand.NewSource(cfg.Seed))
	out := make([]ingestion.Record, 0, cfg.Steps*4)
	for step := 0; step < cfg.Steps; step++ {
		eventTime := cfg.BaseTime.Add(time.Duration(step) * time.Minute)
		tradePrice := rounded(185 + rng.NormFloat64()*0.8)
		bid := rounded(tradePrice - 0.02 - rng.Float64()*0.03)
		ask := rounded(tradePrice + 0.02 + rng.Float64()*0.03)
		windActual := rounded(620 + 35*math.Sin(float64(step)/3) + rng.NormFloat64()*8)
		windForecast := rounded(windActual + rng.NormFloat64()*18)

		payloads := []Payload{
			{Type: "trade", SeriesID: SeriesAppleTradePrice, EventTime: eventTime, Symbol: "AAPL", Price: tradePrice, Quantity: rounded(100 + rng.Float64()*900), Unit: "USD"},
			{Type: "quote", SeriesID: SeriesAppleQuote, EventTime: eventTime, Symbol: "AAPL", Bid: bid, Ask: ask, Quantity: rounded(500 + rng.Float64()*1500), Unit: "USD"},
			{Type: "forecast", SeriesID: SeriesDK1WindForecast, EventTime: eventTime, BiddingZone: "DK1", Value: windForecast, Unit: "MW", ForecastRunID: "synthetic-wind-run-001", IssuedAt: cfg.BaseTime, TargetTime: eventTime.Add(30 * time.Minute), HorizonMinutes: 30 + step},
			{Type: "actual", SeriesID: SeriesDK1WindActual, EventTime: eventTime, BiddingZone: "DK1", Value: windActual, Unit: "MW"},
		}
		for _, payload := range payloads {
			payload.Sequence = fmt.Sprintf("%06d", len(out)+1)
			record, err := record(payload)
			if err != nil {
				return nil, err
			}
			out = append(out, record)
		}
	}
	return out, nil
}

func Adapter(cfg Config) (*ingestion.FakeAdapter, error) {
	records, err := Records(cfg)
	if err != nil {
		return nil, err
	}
	return &ingestion.FakeAdapter{AdapterName: SourceID, Records: records}, nil
}

type Parser struct{}

func (Parser) Parse(_ context.Context, record ingestion.Record) (ingestion.CanonicalEvent, error) {
	var payload Payload
	if err := json.Unmarshal(record.Payload, &payload); err != nil {
		return ingestion.CanonicalEvent{}, err
	}
	if payload.SeriesID == "" || payload.Type == "" {
		return ingestion.CanonicalEvent{}, fmt.Errorf("%w: payload series_id and type are required", ingestion.ErrInvalidRecord)
	}
	return ingestion.CanonicalEvent{
		ID:       record.SourceSequence,
		SeriesID: payload.SeriesID,
		Type:     payload.Type,
		Payload:  append([]byte(nil), record.Payload...),
		Metadata: map[string]string{"provider_id": record.ProviderID, "source_id": record.SourceID},
	}, nil
}

func defaults(cfg Config) Config {
	if cfg.Seed == 0 {
		cfg.Seed = 1
	}
	if cfg.BaseTime.IsZero() {
		cfg.BaseTime = time.Unix(1704067200, 0).UTC()
	}
	if cfg.Steps <= 0 {
		cfg.Steps = 1
	}
	return cfg
}

func record(payload Payload) (ingestion.Record, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return ingestion.Record{}, err
	}
	return ingestion.Record{
		ProviderID:     ProviderID,
		SourceID:       SourceID,
		SourceSequence: payload.Sequence,
		ReceiveTime:    payload.EventTime.Add(time.Second),
		SourceTime:     payload.EventTime,
		Encoding:       "application/json",
		Payload:        raw,
		Metadata:       map[string]string{"series_id": payload.SeriesID, "record_type": payload.Type},
	}, nil
}

func rounded(v float64) float64 {
	return math.Round(v*100) / 100
}
