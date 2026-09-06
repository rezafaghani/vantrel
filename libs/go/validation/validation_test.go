package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
)

func TestEngineAcceptsValidQuote(t *testing.T) {
	err := Engine{}.Validate(context.Background(), ingestion.CanonicalEvent{
		ID:       "1",
		SeriesID: "SYN_EQ_AAPL_QUOTE",
		Type:     "quote",
		Payload:  []byte(`{"bid":185.10,"ask":185.12}`),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEngineRejectsBidGreaterThanAsk(t *testing.T) {
	err := Engine{}.Validate(context.Background(), ingestion.CanonicalEvent{
		ID:       "1",
		SeriesID: "SYN_EQ_AAPL_QUOTE",
		Type:     "quote",
		Payload:  []byte(`{"bid":185.13,"ask":185.12}`),
	})
	var validationErr Error
	if !errors.As(err, &validationErr) {
		t.Fatalf("err=%v", err)
	}
	if validationErr.Result.Valid || validationErr.Result.Findings[0].Code != "BID_GT_ASK" {
		t.Fatalf("result=%+v", validationErr.Result)
	}
}

func TestValidateReportsStructuralFindings(t *testing.T) {
	result := Validate(context.Background(), ingestion.CanonicalEvent{})
	if result.Valid || len(result.Findings) != 4 {
		t.Fatalf("result=%+v", result)
	}
}

func TestEngineRejectsInvalidJSON(t *testing.T) {
	err := Engine{}.Validate(context.Background(), ingestion.CanonicalEvent{
		ID:       "1",
		SeriesID: "SYN_DK1_WIND_ACTUAL",
		Type:     "actual",
		Payload:  []byte(`{`),
	})
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("err=%v", err)
	}
}
