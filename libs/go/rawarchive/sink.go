package rawarchive

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
	"github.com/rezafaghani/vantrel/libs/go/mrx"
)

const Creator = "vantrel-rawarchive-go"

type Sink struct {
	writer *mrx.Writer
}

func NewSink(w io.Writer, providerID, sourceID string, createdAt time.Time) (*Sink, error) {
	if providerID == "" || sourceID == "" {
		return nil, errors.New("rawarchive: provider id and source id are required")
	}
	writer, err := mrx.NewWriter(w, mrx.Header{
		CreatedAtNS: createdAt.UnixNano(),
		Metadata: map[string]any{
			"creator":     Creator,
			"provider_id": providerID,
			"source_id":   sourceID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &Sink{writer: writer}, nil
}

func (s *Sink) Capture(_ context.Context, record ingestion.Record) error {
	if s == nil || s.writer == nil {
		return errors.New("rawarchive: nil sink")
	}
	if record.ProviderID == "" || record.SourceID == "" || record.Encoding == "" || record.ReceiveTime.IsZero() || len(record.Payload) == 0 {
		return errors.New("rawarchive: provider id, source id, encoding, receive time and payload are required")
	}
	seq, hasSeq := sourceSequence(record.SourceSequence)
	return s.writer.WriteFrame(mrx.Frame{
		RawRecordID:    rawRecordID(record),
		ReceiveTimeNS:  record.ReceiveTime.UnixNano(),
		SourceTimeNS:   sourceTimeNS(record.SourceTime),
		SourceSequence: seq,
		Flags:          flags(record, hasSeq),
		Compression:    mrx.CompressionNone,
		Metadata:       metadata(record),
		Payload:        append([]byte(nil), record.Payload...),
	})
}

func (s *Sink) Close() error {
	if s == nil || s.writer == nil {
		return nil
	}
	return s.writer.Close()
}

func metadata(record ingestion.Record) map[string]any {
	out := make(map[string]any, len(record.Metadata)+4)
	for k, v := range record.Metadata {
		out[k] = v
	}
	out["provider_id"] = record.ProviderID
	out["source_id"] = record.SourceID
	out["encoding"] = record.Encoding
	if record.SourceSequence != "" {
		out["source_sequence"] = record.SourceSequence
	}
	return out
}

func sourceSequence(value string) (int64, bool) {
	if value == "" {
		return -1, false
	}
	seq, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1, false
	}
	return seq, true
}

func sourceTimeNS(t time.Time) int64 {
	if t.IsZero() {
		return -1
	}
	return t.UnixNano()
}

func flags(record ingestion.Record, hasSeq bool) uint32 {
	var out uint32
	if hasSeq {
		out |= 1 << 0
	}
	if !record.SourceTime.IsZero() {
		out |= 1 << 1
	}
	if record.Encoding == "application/json" || record.Encoding == "utf-8" {
		out |= 1 << 2
	}
	return out
}

func rawRecordID(record ingestion.Record) [16]byte {
	h := sha256.New()
	_, _ = h.Write([]byte(record.ProviderID))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(record.SourceID))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(record.SourceSequence))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write(record.Payload)
	sum := h.Sum(nil)
	var id [16]byte
	copy(id[:], sum)
	return id
}
