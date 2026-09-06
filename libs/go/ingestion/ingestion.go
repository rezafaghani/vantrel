package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCheckpointNotFound = errors.New("ingestion: checkpoint not found")
	ErrInvalidRecord      = errors.New("ingestion: invalid record")
)

type Mode string

const (
	ModeLive       Mode = "LIVE"
	ModeHistorical Mode = "HISTORICAL"
	ModeReplay     Mode = "REPLAY"
)

type Record struct {
	ProviderID     string
	SourceID       string
	SourceSequence string
	ReceiveTime    time.Time
	SourceTime     time.Time
	Encoding       string
	Payload        []byte
	Metadata       map[string]string
}

type CanonicalEvent struct {
	ID       string
	SeriesID string
	Type     string
	Payload  []byte
	Metadata map[string]string
}

type Checkpoint struct {
	Adapter string
	Mode    Mode
	Value   string
}

type Adapter interface {
	Name() string
	Open(context.Context, Checkpoint) error
	Next(context.Context) (Record, error)
	Close(context.Context) error
}

type CheckpointStore interface {
	Load(context.Context, string, Mode) (Checkpoint, error)
	Save(context.Context, Checkpoint) error
}

type RawSink interface {
	Capture(context.Context, Record) error
}

type Parser interface {
	Parse(context.Context, Record) (CanonicalEvent, error)
}

type Validator interface {
	Validate(context.Context, CanonicalEvent) error
}

type Publisher interface {
	Publish(context.Context, CanonicalEvent) error
}

type Quarantine interface {
	Quarantine(context.Context, Record, error) error
}

type Hooks struct {
	RawSink    RawSink
	Parser     Parser
	Validator  Validator
	Publisher  Publisher
	Quarantine Quarantine
	Checkpoint CheckpointStore
}

type Runtime struct {
	Adapter Adapter
	Mode    Mode
	Hooks   Hooks
	Retries int
	Backoff time.Duration
}

func (r Runtime) Run(ctx context.Context) (err error) {
	if r.Adapter == nil {
		return errors.New("ingestion: adapter is required")
	}
	if r.Hooks.Parser == nil || r.Hooks.Publisher == nil {
		return errors.New("ingestion: parser and publisher are required")
	}
	if r.Mode == "" {
		r.Mode = ModeLive
	}
	checkpoint := Checkpoint{Adapter: r.Adapter.Name(), Mode: r.Mode}
	if r.Hooks.Checkpoint != nil {
		loaded, err := r.Hooks.Checkpoint.Load(ctx, r.Adapter.Name(), r.Mode)
		if err != nil && !errors.Is(err, ErrCheckpointNotFound) {
			return err
		}
		if err == nil {
			checkpoint = loaded
		}
	}
	if err := r.Adapter.Open(ctx, checkpoint); err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Adapter.Close(ctx); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	for {
		record, err := retry(ctx, r.Retries, r.Backoff, r.Adapter.Next)
		if err != nil {
			if errors.Is(err, EndOfSource) {
				return nil
			}
			return err
		}
		if err := validateRecord(record); err != nil {
			return err
		}
		if err := r.process(ctx, record); err != nil {
			return err
		}
		if r.Hooks.Checkpoint != nil {
			checkpoint.Value = record.SourceSequence
			if err := r.Hooks.Checkpoint.Save(ctx, checkpoint); err != nil {
				return err
			}
		}
	}
}

func (r Runtime) process(ctx context.Context, record Record) error {
	if r.Hooks.RawSink != nil {
		if err := r.Hooks.RawSink.Capture(ctx, record); err != nil {
			return err
		}
	}
	event, err := r.Hooks.Parser.Parse(ctx, record)
	if err != nil {
		return r.quarantine(ctx, record, err)
	}
	if r.Hooks.Validator != nil {
		if err := r.Hooks.Validator.Validate(ctx, event); err != nil {
			return r.quarantine(ctx, record, err)
		}
	}
	return r.Hooks.Publisher.Publish(ctx, event)
}

func (r Runtime) quarantine(ctx context.Context, record Record, cause error) error {
	if r.Hooks.Quarantine == nil {
		return cause
	}
	return r.Hooks.Quarantine.Quarantine(ctx, record, cause)
}

func validateRecord(r Record) error {
	if r.ProviderID == "" || r.SourceID == "" || len(r.Payload) == 0 {
		return fmt.Errorf("%w: provider id, source id and payload are required", ErrInvalidRecord)
	}
	return nil
}

func retry(ctx context.Context, attempts int, backoff time.Duration, fn func(context.Context) (Record, error)) (Record, error) {
	if attempts < 0 {
		attempts = 0
	}
	var last error
	for i := 0; i <= attempts; i++ {
		record, err := fn(ctx)
		if err == nil || errors.Is(err, EndOfSource) {
			return record, err
		}
		last = err
		if backoff > 0 && i < attempts {
			timer := time.NewTimer(backoffForAttempt(backoff, i))
			select {
			case <-ctx.Done():
				timer.Stop()
				return Record{}, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return Record{}, last
}

func backoffForAttempt(base time.Duration, attempt int) time.Duration {
	if attempt <= 0 {
		return base
	}
	return base << min(attempt, 10)
}

var EndOfSource = errors.New("ingestion: end of source")
