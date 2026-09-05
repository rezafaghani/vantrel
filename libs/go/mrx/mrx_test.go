package mrx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"testing"
)

func TestRoundTripMultipleFrames(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, testHeader())
	if err != nil {
		t.Fatal(err)
	}
	payloads := [][]byte{
		[]byte(`{"price":42.5,"symbol":"AAPL"}`),
		[]byte("SEQ=100\n"),
	}
	for i, payload := range payloads {
		if err := w.WriteFrame(Frame{
			RawRecordID:    id(byte(i + 1)),
			ReceiveTimeNS:  1735689600000000000 + int64(i),
			SourceTimeNS:   1735689599000000000 + int64(i),
			SourceSequence: int64(100 + i),
			Flags:          7,
			Compression:    CompressionNone,
			Metadata:       frameMetadata(),
			Payload:        payload,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	a := Read(bytes.NewReader(buf.Bytes()))
	if a.Status != StatusClean || a.Err != nil {
		t.Fatalf("status=%s err=%v", a.Status, a.Err)
	}
	if len(a.Frames) != len(payloads) || len(a.Index) != len(payloads) {
		t.Fatalf("frames=%d index=%d", len(a.Frames), len(a.Index))
	}
	for i, want := range payloads {
		if got := a.Frames[i].Payload; !bytes.Equal(got, want) {
			t.Fatalf("frame %d payload = %q, want %q", i, got, want)
		}
		if a.Frames[i].FrameNumber != uint64(i) {
			t.Fatalf("frame number = %d, want %d", a.Frames[i].FrameNumber, i)
		}
	}
	f, ok := a.FrameByNumber(1)
	if !ok || !bytes.Equal(f.Payload, payloads[1]) {
		t.Fatalf("seek by frame number failed")
	}
}

func TestChecksumVectors(t *testing.T) {
	cases := map[string]uint32{
		`{"price":42.5,"symbol":"AAPL"}`: 0xc26d4ad2,
		"SEQ=100\n":                      0x8f81a4e8,
	}
	for payload, want := range cases {
		if got := crc32.Checksum([]byte(payload), castagnoli); got != want {
			t.Fatalf("crc32c(%q)=%08x, want %08x", payload, got, want)
		}
	}
}

func TestCorruption(t *testing.T) {
	data := oneFrameArchive(t)
	data[len(data)-40] ^= 0xff

	a := Read(bytes.NewReader(data))
	if a.Status != StatusCorrupt || !errors.Is(a.Err, ErrChecksum) {
		t.Fatalf("status=%s err=%v", a.Status, a.Err)
	}
}

func TestTruncatedFileRecoversCompleteFrame(t *testing.T) {
	data := oneFrameArchive(t)
	data = data[:len(data)-36]

	a := Read(bytes.NewReader(data))
	if a.Status != StatusUncleanRecoverable || !errors.Is(a.Err, ErrTruncated) {
		t.Fatalf("status=%s err=%v", a.Status, a.Err)
	}
	if len(a.Frames) != 1 || string(a.Frames[0].Payload) != "SEQ=100\n" {
		t.Fatalf("frames=%v", a.Frames)
	}
}

func TestUnsupportedCompression(t *testing.T) {
	data := oneFrameArchive(t)
	frameOffset := int(binary.LittleEndian.Uint16(data[6:8]))
	data[frameOffset+7] = byte(CompressionZSTD)
	binary.LittleEndian.PutUint32(data[frameOffset+84:frameOffset+88], crc32.Checksum(data[frameOffset:frameOffset+84], castagnoli))

	a := Read(bytes.NewReader(data))
	if a.Status != StatusEmptyOrUnreadable || !errors.Is(a.Err, ErrUnsupportedCompression) {
		t.Fatalf("status=%s err=%v", a.Status, a.Err)
	}
}

func TestWriterRejectsUnsupportedCompression(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, testHeader())
	if err != nil {
		t.Fatal(err)
	}
	err = w.WriteFrame(Frame{Compression: CompressionZSTD, Metadata: frameMetadata()})
	if !errors.Is(err, ErrUnsupportedCompression) {
		t.Fatalf("err=%v", err)
	}
}

func oneFrameArchive(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := NewWriter(&buf, testHeader())
	if err != nil {
		t.Fatal(err)
	}
	if err := w.WriteFrame(Frame{
		RawRecordID:    id(1),
		ReceiveTimeNS:  1735689600000000000,
		SourceTimeNS:   1735689599000000000,
		SourceSequence: 100,
		Flags:          7,
		Compression:    CompressionNone,
		Metadata:       frameMetadata(),
		Payload:        []byte("SEQ=100\n"),
	}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.Clone(buf.Bytes())
}

func testHeader() Header {
	return Header{
		FileID:      id(9),
		CreatedAtNS: 1735689600000000000,
		Metadata: map[string]any{
			"creator":     "mrx-test",
			"provider_id": "test-provider",
			"source_id":   "test-source",
		},
	}
}

func frameMetadata() map[string]any {
	return map[string]any{
		"encoding":    "utf-8",
		"provider_id": "test-provider",
		"source_id":   "test-source",
	}
}

func id(b byte) [16]byte {
	var out [16]byte
	for i := range out {
		out[i] = b
	}
	return out
}
