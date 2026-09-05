package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rezafaghani/vantrel/libs/go/mrx"
)

func TestInspectVerifyExtract(t *testing.T) {
	path := writeArchive(t)

	var inspectOut bytes.Buffer
	if err := run([]string{"inspect", path}, &inspectOut, io.Discard); err != nil {
		t.Fatal(err)
	}
	var summary struct {
		Status string `json:"status"`
		Frames int    `json:"frames"`
	}
	if err := json.Unmarshal(inspectOut.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Status != string(mrx.StatusClean) || summary.Frames != 2 {
		t.Fatalf("summary=%+v", summary)
	}

	var verifyOut bytes.Buffer
	if err := run([]string{"verify", path}, &verifyOut, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := verifyOut.String(); !strings.Contains(got, "CLEAN frames=2") {
		t.Fatalf("verify output=%q", got)
	}

	var extractOut bytes.Buffer
	if err := run([]string{"extract", "-frame", "1", path}, &extractOut, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := extractOut.String(); got != "two" {
		t.Fatalf("extract=%q", got)
	}
}

func TestVerifyFailsCorruptArchive(t *testing.T) {
	path := writeArchive(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-40] ^= 0xff
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := run([]string{"verify", path}, &out, io.Discard); err == nil {
		t.Fatal("verify succeeded for corrupt archive")
	}
}

func TestBenchmark(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"benchmark", "-frames", "2", "-size", "8"}, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "frames=2 payload_bytes=16") {
		t.Fatalf("benchmark output=%q", got)
	}
}

func writeArchive(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.mrx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w, err := mrx.NewWriter(f, mrx.Header{
		CreatedAtNS: 1,
		Metadata: map[string]any{
			"creator":     "mrx-cli-test",
			"provider_id": "test",
			"source_id":   "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range [][]byte{[]byte("one"), []byte("two")} {
		if err := w.WriteFrame(mrx.Frame{
			ReceiveTimeNS: 1,
			SourceTimeNS:  -1,
			Compression:   mrx.CompressionNone,
			Metadata: map[string]any{
				"encoding":    "utf-8",
				"provider_id": "test",
				"source_id":   "test",
			},
			Payload: payload,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
