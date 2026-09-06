package validation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
)

var ErrInvalidEvent = errors.New("validation: invalid event")

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
)

type Category string

const (
	CategoryStructural Category = "STRUCTURAL"
	CategorySemantic   Category = "SEMANTIC"
)

type Finding struct {
	Code     string
	Category Category
	Severity Severity
	Message  string
}

type Result struct {
	EventID  string
	SeriesID string
	Valid    bool
	Findings []Finding
}

type Engine struct{}

func (Engine) Validate(ctx context.Context, event ingestion.CanonicalEvent) error {
	result := Validate(ctx, event)
	if result.Valid {
		return nil
	}
	return Error{Result: result}
}

type Error struct {
	Result Result
}

func (e Error) Error() string {
	return fmt.Sprintf("%v: %d finding(s)", ErrInvalidEvent, len(e.Result.Findings))
}

func (e Error) Unwrap() error {
	return ErrInvalidEvent
}

func Validate(_ context.Context, event ingestion.CanonicalEvent) Result {
	result := Result{EventID: event.ID, SeriesID: event.SeriesID, Valid: true}
	add := func(code string, category Category, message string) {
		result.Valid = false
		result.Findings = append(result.Findings, Finding{Code: code, Category: category, Severity: SeverityError, Message: message})
	}
	if event.ID == "" {
		add("MISSING_EVENT_ID", CategoryStructural, "event id is required")
	}
	if event.SeriesID == "" {
		add("MISSING_SERIES_ID", CategoryStructural, "series id is required")
	}
	if event.Type == "" {
		add("MISSING_TYPE", CategoryStructural, "event type is required")
	}
	if len(event.Payload) == 0 {
		add("MISSING_PAYLOAD", CategoryStructural, "payload is required")
	}
	if len(event.Payload) == 0 {
		return result
	}

	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		add("INVALID_JSON_PAYLOAD", CategoryStructural, "payload must be JSON for initial validation")
		return result
	}
	switch event.Type {
	case "quote":
		validateQuote(payload, add)
	case "actual", "forecast":
		validateValue(payload, add)
	}
	return result
}

func validateQuote(payload map[string]any, add func(string, Category, string)) {
	bid, bidOK := number(payload["bid"])
	ask, askOK := number(payload["ask"])
	if !bidOK || !askOK {
		add("MISSING_BID_ASK", CategoryStructural, "quote payload requires numeric bid and ask")
		return
	}
	if bid > ask {
		add("BID_GT_ASK", CategorySemantic, "bid must be less than or equal to ask")
	}
}

func validateValue(payload map[string]any, add func(string, Category, string)) {
	value, ok := number(payload["value"])
	if !ok {
		add("MISSING_VALUE", CategoryStructural, "observation payload requires numeric value")
		return
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		add("INVALID_VALUE", CategorySemantic, "observation value must be finite")
	}
}

func number(v any) (float64, bool) {
	n, ok := v.(float64)
	return n, ok
}
