package marketstore

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

var (
	ErrInvalid = errors.New("marketstore: invalid observation")
)

type Lineage struct {
	RawRecordID      string
	CanonicalEventID string
	IngestionID      string
}

type Quality struct {
	State string
	Score float64
	Flags []string
}

type ActualObservation struct {
	SeriesID  string
	EventTime time.Time
	Value     float64
	Unit      string
	Quality   Quality
	Revision  uint64
	Lineage   Lineage
}

type ForecastObservation struct {
	SeriesID      string
	ForecastRunID string
	IssuedAt      time.Time
	TargetTime    time.Time
	Horizon       time.Duration
	Value         float64
	Unit          string
	ModelVersion  string
	Quality       Quality
	Lineage       Lineage
}

type Query struct {
	SeriesID string
	From     time.Time
	To       time.Time
}

type Store interface {
	ObservationWriter
	QueryActual(Query) ([]ActualObservation, error)
	QueryForecast(Query) ([]ForecastObservation, error)
}

type ObservationWriter interface {
	WriteActual(ActualObservation) error
	WriteForecast(ForecastObservation) error
}

type MemoryStore struct {
	mu        sync.RWMutex
	actuals   []ActualObservation
	forecasts []ForecastObservation
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) WriteActual(obs ActualObservation) error {
	if err := ValidateActual(obs); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actuals = append(s.actuals, cloneActual(obs))
	return nil
}

func (s *MemoryStore) WriteForecast(obs ForecastObservation) error {
	if err := ValidateForecast(obs); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forecasts = append(s.forecasts, cloneForecast(obs))
	return nil
}

func (s *MemoryStore) QueryActual(q Query) ([]ActualObservation, error) {
	if err := ValidateQuery(q); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ActualObservation
	for _, obs := range s.actuals {
		if obs.SeriesID == q.SeriesID && inRange(obs.EventTime, q.From, q.To) {
			out = append(out, cloneActual(obs))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].EventTime.Before(out[j].EventTime)
	})
	return out, nil
}

func (s *MemoryStore) QueryForecast(q Query) ([]ForecastObservation, error) {
	if err := ValidateQuery(q); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ForecastObservation
	for _, obs := range s.forecasts {
		if obs.SeriesID == q.SeriesID && inRange(obs.TargetTime, q.From, q.To) {
			out = append(out, cloneForecast(obs))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TargetTime.Before(out[j].TargetTime)
	})
	return out, nil
}

func ValidateActual(obs ActualObservation) error {
	if obs.SeriesID == "" || obs.EventTime.IsZero() || obs.Unit == "" {
		return fmt.Errorf("%w: actual series id, event time and unit are required", ErrInvalid)
	}
	if !finite(obs.Value) {
		return fmt.Errorf("%w: actual value must be finite", ErrInvalid)
	}
	return nil
}

func ValidateForecast(obs ForecastObservation) error {
	if obs.SeriesID == "" || obs.ForecastRunID == "" || obs.IssuedAt.IsZero() || obs.TargetTime.IsZero() || obs.Unit == "" {
		return fmt.Errorf("%w: forecast series id, run id, issued time, target time and unit are required", ErrInvalid)
	}
	if obs.Horizon < 0 {
		return fmt.Errorf("%w: forecast horizon cannot be negative", ErrInvalid)
	}
	if !finite(obs.Value) {
		return fmt.Errorf("%w: forecast value must be finite", ErrInvalid)
	}
	return nil
}

func ValidateQuery(q Query) error {
	if q.SeriesID == "" {
		return fmt.Errorf("%w: series id is required", ErrInvalid)
	}
	if !q.From.IsZero() && !q.To.IsZero() && q.From.After(q.To) {
		return fmt.Errorf("%w: from cannot be after to", ErrInvalid)
	}
	return nil
}

func inRange(t, from, to time.Time) bool {
	if !from.IsZero() && t.Before(from) {
		return false
	}
	return to.IsZero() || !t.After(to)
}

func cloneActual(obs ActualObservation) ActualObservation {
	obs.Quality.Flags = append([]string(nil), obs.Quality.Flags...)
	return obs
}

func cloneForecast(obs ForecastObservation) ForecastObservation {
	obs.Quality.Flags = append([]string(nil), obs.Quality.Flags...)
	return obs
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
