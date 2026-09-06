package ingestion

import (
	"context"
	"errors"
	"testing"
)

func TestRuntimeProcessesSharedPipeline(t *testing.T) {
	ctx := context.Background()
	raw := &recordSink{}
	parser := parserFunc(func(_ context.Context, record Record) (CanonicalEvent, error) {
		return CanonicalEvent{ID: record.SourceSequence, SeriesID: "dk1_wind_actual", Type: "actual", Payload: record.Payload}, nil
	})
	validator := validatorFunc(func(_ context.Context, event CanonicalEvent) error {
		if event.SeriesID == "" {
			return errors.New("missing series")
		}
		return nil
	})
	publisher := &eventSink{}
	checkpoints := NewMemoryCheckpointStore()

	err := Runtime{
		Adapter: &FakeAdapter{Records: []Record{
			record("1", "one"),
			record("2", "two"),
		}},
		Mode: ModeHistorical,
		Hooks: Hooks{
			RawSink:    raw,
			Parser:     parser,
			Validator:  validator,
			Publisher:  publisher,
			Checkpoint: checkpoints,
		},
	}.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw.records) != 2 || len(publisher.events) != 2 {
		t.Fatalf("raw=%d events=%d", len(raw.records), len(publisher.events))
	}
	cp, err := checkpoints.Load(ctx, "fake", ModeHistorical)
	if err != nil {
		t.Fatal(err)
	}
	if cp.Value != "2" {
		t.Fatalf("checkpoint=%+v", cp)
	}
}

func TestRuntimeQuarantinesInvalidCanonicalEvent(t *testing.T) {
	ctx := context.Background()
	quarantine := &quarantineSink{}
	err := Runtime{
		Adapter: &FakeAdapter{Records: []Record{record("1", "bad")}},
		Hooks: Hooks{
			Parser: parserFunc(func(context.Context, Record) (CanonicalEvent, error) {
				return CanonicalEvent{ID: "bad"}, nil
			}),
			Validator: validatorFunc(func(context.Context, CanonicalEvent) error {
				return errors.New("invalid canonical event")
			}),
			Publisher:  &eventSink{},
			Quarantine: quarantine,
		},
	}.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(quarantine.records) != 1 {
		t.Fatalf("quarantined=%d", len(quarantine.records))
	}
}

func TestRuntimeRetriesAdapterNext(t *testing.T) {
	err := Runtime{
		Adapter: &FakeAdapter{Failures: 1, Records: []Record{record("1", "one")}},
		Retries: 1,
		Hooks: Hooks{
			Parser: parserFunc(func(_ context.Context, r Record) (CanonicalEvent, error) {
				return CanonicalEvent{ID: r.SourceSequence, Payload: r.Payload}, nil
			}),
			Publisher: &eventSink{},
		},
	}.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeResumesFromCheckpoint(t *testing.T) {
	ctx := context.Background()
	checkpoints := NewMemoryCheckpointStore()
	must(t, checkpoints.Save(ctx, Checkpoint{Adapter: "fake", Mode: ModeReplay, Value: "1"}))
	publisher := &eventSink{}

	err := Runtime{
		Adapter: &FakeAdapter{Records: []Record{record("1", "one"), record("2", "two")}},
		Mode:    ModeReplay,
		Hooks: Hooks{
			Parser: parserFunc(func(_ context.Context, r Record) (CanonicalEvent, error) {
				return CanonicalEvent{ID: r.SourceSequence, Payload: r.Payload}, nil
			}),
			Publisher:  publisher,
			Checkpoint: checkpoints,
		},
	}.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(publisher.events) != 1 || publisher.events[0].ID != "2" {
		t.Fatalf("events=%+v", publisher.events)
	}
}

func TestRuntimeRequiresCoreHooks(t *testing.T) {
	err := Runtime{Adapter: &FakeAdapter{}}.Run(context.Background())
	if err == nil {
		t.Fatal("expected missing hook error")
	}
}

func TestBackoffForAttempt(t *testing.T) {
	if got := backoffForAttempt(2, 3); got != 16 {
		t.Fatalf("backoff=%s", got)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func record(sequence, payload string) Record {
	return Record{
		ProviderID:     "test-provider",
		SourceID:       "test-source",
		SourceSequence: sequence,
		Encoding:       "utf-8",
		Payload:        []byte(payload),
	}
}

type recordSink struct {
	records []Record
}

func (s *recordSink) Capture(_ context.Context, r Record) error {
	s.records = append(s.records, r)
	return nil
}

type eventSink struct {
	events []CanonicalEvent
}

func (s *eventSink) Publish(_ context.Context, e CanonicalEvent) error {
	s.events = append(s.events, e)
	return nil
}

type quarantineSink struct {
	records []Record
}

func (s *quarantineSink) Quarantine(_ context.Context, r Record, _ error) error {
	s.records = append(s.records, r)
	return nil
}

type parserFunc func(context.Context, Record) (CanonicalEvent, error)

func (f parserFunc) Parse(ctx context.Context, r Record) (CanonicalEvent, error) {
	return f(ctx, r)
}

type validatorFunc func(context.Context, CanonicalEvent) error

func (f validatorFunc) Validate(ctx context.Context, e CanonicalEvent) error {
	return f(ctx, e)
}
