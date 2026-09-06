package rawarchive

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
	"github.com/rezafaghani/vantrel/libs/go/mrx"
)

func TestSinkArchivesRecordsBytePerfectly(t *testing.T) {
	records := testRecords()
	var buf bytes.Buffer
	sink, err := NewSink(&buf, "synthetic", "synthetic-market-source", time.Unix(1704067200, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if err := sink.Capture(context.Background(), record); err != nil {
			t.Fatal(err)
		}
	}
	if err := sink.Close(); err != nil {
		t.Fatal(err)
	}

	archive := mrx.Read(bytes.NewReader(buf.Bytes()))
	if archive.Status != mrx.StatusClean {
		t.Fatalf("status=%s err=%v", archive.Status, archive.Err)
	}
	if len(archive.Frames) != len(records) {
		t.Fatalf("frames=%d records=%d", len(archive.Frames), len(records))
	}
	for i, frame := range archive.Frames {
		if string(frame.Payload) != string(records[i].Payload) {
			t.Fatalf("payload %d differs", i)
		}
		if frame.Metadata["series_id"] != records[i].Metadata["series_id"] {
			t.Fatalf("metadata=%v", frame.Metadata)
		}
	}
}

func testRecords() []ingestion.Record {
	return []ingestion.Record{
		{
			ProviderID:     "synthetic",
			SourceID:       "synthetic-market-source",
			SourceSequence: "1",
			ReceiveTime:    time.Unix(1704067201, 0).UTC(),
			SourceTime:     time.Unix(1704067200, 0).UTC(),
			Encoding:       "application/json",
			Payload:        []byte(`{"type":"trade","series_id":"SYN_EQ_AAPL_TRADE_PRICE","price":185.1}`),
			Metadata:       map[string]string{"series_id": "SYN_EQ_AAPL_TRADE_PRICE"},
		},
		{
			ProviderID:     "synthetic",
			SourceID:       "synthetic-market-source",
			SourceSequence: "2",
			ReceiveTime:    time.Unix(1704067261, 0).UTC(),
			SourceTime:     time.Unix(1704067260, 0).UTC(),
			Encoding:       "application/json",
			Payload:        []byte(`{"type":"actual","series_id":"SYN_DK1_WIND_ACTUAL","value":620.2}`),
			Metadata:       map[string]string{"series_id": "SYN_DK1_WIND_ACTUAL"},
		},
	}
}

func TestSinkRejectsMissingRequiredFields(t *testing.T) {
	var buf bytes.Buffer
	sink, err := NewSink(&buf, "provider", "source", time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	err = sink.Capture(context.Background(), ingestion.Record{ProviderID: "provider", SourceID: "source", Payload: []byte("raw")})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSinkPreservesNonNumericSequenceInMetadata(t *testing.T) {
	var buf bytes.Buffer
	sink, err := NewSink(&buf, "provider", "source", time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	err = sink.Capture(context.Background(), ingestion.Record{
		ProviderID:     "provider",
		SourceID:       "source",
		SourceSequence: "abc",
		ReceiveTime:    time.Unix(1, 0),
		Encoding:       "application/json",
		Payload:        []byte("{}"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Close(); err != nil {
		t.Fatal(err)
	}
	archive := mrx.Read(bytes.NewReader(buf.Bytes()))
	if archive.Status != mrx.StatusClean {
		t.Fatalf("status=%s err=%v", archive.Status, archive.Err)
	}
	if archive.Frames[0].SourceSequence != -1 || archive.Frames[0].Metadata["source_sequence"] != "abc" {
		t.Fatalf("frame=%+v", archive.Frames[0])
	}
}
