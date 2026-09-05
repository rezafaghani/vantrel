package mrx

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"math"
)

var (
	fileMagic      = [4]byte{'M', 'R', 'X', '0'}
	frameMagic     = [4]byte{'M', 'R', 'X', 'F'}
	footerMagic    = [4]byte{'M', 'R', 'X', 'E'}
	footerEndMagic = [4]byte{'D', 'O', 'N', 'E'}

	castagnoli = crc32.MakeTable(crc32.Castagnoli)

	ErrUnsupportedCompression = errors.New("mrx: unsupported compression")
	ErrChecksum               = errors.New("mrx: checksum mismatch")
	ErrInvalidMagic           = errors.New("mrx: invalid magic")
	ErrTruncated              = errors.New("mrx: truncated archive")
)

type CompressionID uint8

const (
	CompressionNone CompressionID = 0
	CompressionLZ4  CompressionID = 1
	CompressionZSTD CompressionID = 2
)

type Status string

const (
	StatusClean              Status = "CLEAN"
	StatusUncleanRecoverable Status = "UNCLEAN_RECOVERABLE"
	StatusEmptyOrUnreadable  Status = "EMPTY_OR_UNREADABLE"
	StatusCorrupt            Status = "CORRUPT"
)

type Header struct {
	FileID      [16]byte
	CreatedAtNS int64
	Metadata    map[string]any
}

type Frame struct {
	FrameNumber    uint64
	RawRecordID    [16]byte
	ReceiveTimeNS  int64
	SourceTimeNS   int64
	SourceSequence int64
	Flags          uint32
	Compression    CompressionID
	Metadata       map[string]any
	Payload        []byte
}

type FrameIndex struct {
	FrameNumber    uint64
	Offset         uint64
	ReceiveTimeNS  int64
	SourceTimeNS   int64
	SourceSequence int64
}

type Archive struct {
	Header Header
	Frames []Frame
	Index  []FrameIndex
	Status Status
	Err    error
}

type Writer struct {
	w          io.Writer
	crc        hash.Hash32
	frameCount uint64
	closed     bool
}

func NewWriter(w io.Writer, h Header) (*Writer, error) {
	if w == nil {
		return nil, errors.New("mrx: nil writer")
	}
	if err := requireMetadata(h.Metadata, "creator", "provider_id", "source_id"); err != nil {
		return nil, err
	}
	metadata, err := canonicalJSON(h.Metadata)
	if err != nil {
		return nil, err
	}
	if len(metadata) > math.MaxUint32 {
		return nil, errors.New("mrx: header metadata too large")
	}
	headerLen := 40 + len(metadata)
	if headerLen > math.MaxUint16 {
		return nil, errors.New("mrx: header too large")
	}

	buf := make([]byte, headerLen)
	copy(buf[0:4], fileMagic[:])
	buf[4] = 0
	buf[5] = 1
	binary.LittleEndian.PutUint16(buf[6:8], uint16(headerLen))
	binary.LittleEndian.PutUint32(buf[8:12], 1)
	copy(buf[12:28], h.FileID[:])
	binary.LittleEndian.PutUint64(buf[28:36], uint64(h.CreatedAtNS))
	binary.LittleEndian.PutUint32(buf[36:40], uint32(len(metadata)))
	copy(buf[40:], metadata)

	c := crc32.New(castagnoli)
	if _, err := w.Write(buf); err != nil {
		return nil, err
	}
	_, _ = c.Write(buf)
	return &Writer{w: w, crc: c}, nil
}

func (w *Writer) WriteFrame(f Frame) error {
	if w.closed {
		return errors.New("mrx: writer closed")
	}
	if f.Compression != CompressionNone {
		return fmt.Errorf("%w: %d", ErrUnsupportedCompression, f.Compression)
	}
	if err := requireMetadata(f.Metadata, "provider_id", "source_id", "encoding"); err != nil {
		return err
	}
	metadata, err := canonicalJSON(f.Metadata)
	if err != nil {
		return err
	}
	if len(metadata) > math.MaxUint32 {
		return errors.New("mrx: frame metadata too large")
	}
	headerLen := 88 + len(metadata)
	if headerLen > math.MaxUint16 {
		return errors.New("mrx: frame header too large")
	}

	header := make([]byte, headerLen)
	copy(header[0:4], frameMagic[:])
	binary.LittleEndian.PutUint16(header[4:6], uint16(headerLen))
	header[6] = 0
	header[7] = byte(f.Compression)
	binary.LittleEndian.PutUint32(header[8:12], f.Flags)
	binary.LittleEndian.PutUint64(header[12:20], w.frameCount)
	copy(header[20:36], f.RawRecordID[:])
	binary.LittleEndian.PutUint64(header[36:44], uint64(f.ReceiveTimeNS))
	binary.LittleEndian.PutUint64(header[44:52], uint64(f.SourceTimeNS))
	binary.LittleEndian.PutUint64(header[52:60], uint64(f.SourceSequence))
	binary.LittleEndian.PutUint32(header[60:64], uint32(len(metadata)))
	binary.LittleEndian.PutUint64(header[64:72], uint64(len(f.Payload)))
	binary.LittleEndian.PutUint64(header[72:80], uint64(len(f.Payload)))
	binary.LittleEndian.PutUint32(header[80:84], crc32.Checksum(f.Payload, castagnoli))
	binary.LittleEndian.PutUint32(header[84:88], crc32.Checksum(header[:84], castagnoli))
	copy(header[88:], metadata)

	if err := w.writeCounted(header); err != nil {
		return err
	}
	if err := w.writeCounted(f.Payload); err != nil {
		return err
	}
	w.frameCount++
	return nil
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true
	footer := make([]byte, 36)
	copy(footer[0:4], footerMagic[:])
	binary.LittleEndian.PutUint64(footer[4:12], w.frameCount)
	binary.LittleEndian.PutUint64(footer[12:20], 0)
	binary.LittleEndian.PutUint64(footer[20:28], 0)
	binary.LittleEndian.PutUint32(footer[28:32], w.crc.Sum32())
	copy(footer[32:36], footerEndMagic[:])
	_, err := w.w.Write(footer)
	return err
}

func (w *Writer) writeCounted(p []byte) error {
	if _, err := w.w.Write(p); err != nil {
		return err
	}
	_, _ = w.crc.Write(p)
	return nil
}

func Read(r io.Reader) Archive {
	data, err := io.ReadAll(r)
	if err != nil {
		return Archive{Status: StatusEmptyOrUnreadable, Err: err}
	}
	// ponytail: whole-file read; switch to streaming when archive sizes make memory measurable.
	a, offset, err := readHeader(data)
	if err != nil {
		return Archive{Status: StatusEmptyOrUnreadable, Err: err}
	}

	footerAt, footerOK := findFooter(data)
	end := len(data)
	if footerOK {
		end = footerAt
		if got := crc32.Checksum(data[:footerAt], castagnoli); got != binary.LittleEndian.Uint32(data[footerAt+28:footerAt+32]) {
			a.Status = StatusCorrupt
			a.Err = fmt.Errorf("%w: file", ErrChecksum)
			return a
		}
	}

	for offset < end {
		frameOffset := offset
		f, next, err := readFrame(data, offset, end)
		if err != nil {
			if errors.Is(err, ErrTruncated) && len(a.Frames) > 0 {
				a.Status = StatusUncleanRecoverable
				a.Err = err
				return a
			}
			if len(a.Frames) == 0 {
				a.Status = StatusEmptyOrUnreadable
			} else {
				a.Status = StatusCorrupt
			}
			a.Err = err
			return a
		}
		a.Index = append(a.Index, FrameIndex{
			FrameNumber:    f.FrameNumber,
			Offset:         uint64(frameOffset),
			ReceiveTimeNS:  f.ReceiveTimeNS,
			SourceTimeNS:   f.SourceTimeNS,
			SourceSequence: f.SourceSequence,
		})
		a.Frames = append(a.Frames, f)
		offset = next
	}

	if footerOK {
		if binary.LittleEndian.Uint64(data[footerAt+4:footerAt+12]) != uint64(len(a.Frames)) {
			a.Status = StatusCorrupt
			a.Err = errors.New("mrx: footer frame count mismatch")
			return a
		}
		a.Status = StatusClean
		return a
	}
	if len(a.Frames) == 0 {
		a.Status = StatusEmptyOrUnreadable
		a.Err = ErrTruncated
		return a
	}
	a.Status = StatusUncleanRecoverable
	a.Err = ErrTruncated
	return a
}

func (a Archive) FrameByNumber(n uint64) (Frame, bool) {
	for _, f := range a.Frames {
		if f.FrameNumber == n {
			return f, true
		}
	}
	return Frame{}, false
}

func readHeader(data []byte) (Archive, int, error) {
	if len(data) < 40 {
		return Archive{}, 0, ErrTruncated
	}
	if !bytes.Equal(data[0:4], fileMagic[:]) {
		return Archive{}, 0, ErrInvalidMagic
	}
	if data[4] != 0 {
		return Archive{}, 0, errors.New("mrx: unsupported major version")
	}
	headerLen := int(binary.LittleEndian.Uint16(data[6:8]))
	metadataLen := int(binary.LittleEndian.Uint32(data[36:40]))
	if headerLen != 40+metadataLen || headerLen > len(data) {
		return Archive{}, 0, ErrTruncated
	}
	var metadata map[string]any
	if err := json.Unmarshal(data[40:headerLen], &metadata); err != nil {
		return Archive{}, 0, err
	}
	var fileID [16]byte
	copy(fileID[:], data[12:28])
	return Archive{Header: Header{
		FileID:      fileID,
		CreatedAtNS: int64(binary.LittleEndian.Uint64(data[28:36])),
		Metadata:    metadata,
	}}, headerLen, nil
}

func readFrame(data []byte, offset, end int) (Frame, int, error) {
	if end-offset < 88 {
		return Frame{}, offset, ErrTruncated
	}
	header := data[offset : offset+88]
	if !bytes.Equal(header[0:4], frameMagic[:]) {
		return Frame{}, offset, ErrInvalidMagic
	}
	if got, want := crc32.Checksum(header[:84], castagnoli), binary.LittleEndian.Uint32(header[84:88]); got != want {
		return Frame{}, offset, fmt.Errorf("%w: frame header", ErrChecksum)
	}
	compression := CompressionID(header[7])
	if compression != CompressionNone {
		return Frame{}, offset, fmt.Errorf("%w: %d", ErrUnsupportedCompression, compression)
	}
	headerLen := int(binary.LittleEndian.Uint16(header[4:6]))
	metadataLen := int(binary.LittleEndian.Uint32(header[60:64]))
	if headerLen != 88+metadataLen || offset+headerLen > end {
		return Frame{}, offset, ErrTruncated
	}
	uncompressedLen := binary.LittleEndian.Uint64(header[64:72])
	compressedLen := binary.LittleEndian.Uint64(header[72:80])
	if compressedLen != uncompressedLen {
		return Frame{}, offset, errors.New("mrx: NONE frame length mismatch")
	}
	if compressedLen > uint64(end-offset-headerLen) {
		return Frame{}, offset, ErrTruncated
	}
	payloadLen := int(compressedLen)
	var metadata map[string]any
	if err := json.Unmarshal(data[offset+88:offset+headerLen], &metadata); err != nil {
		return Frame{}, offset, err
	}
	payload := bytes.Clone(data[offset+headerLen : offset+headerLen+payloadLen])
	if got, want := crc32.Checksum(payload, castagnoli), binary.LittleEndian.Uint32(header[80:84]); got != want {
		return Frame{}, offset, fmt.Errorf("%w: payload", ErrChecksum)
	}
	var rawRecordID [16]byte
	copy(rawRecordID[:], header[20:36])
	return Frame{
		FrameNumber:    binary.LittleEndian.Uint64(header[12:20]),
		RawRecordID:    rawRecordID,
		ReceiveTimeNS:  int64(binary.LittleEndian.Uint64(header[36:44])),
		SourceTimeNS:   int64(binary.LittleEndian.Uint64(header[44:52])),
		SourceSequence: int64(binary.LittleEndian.Uint64(header[52:60])),
		Flags:          binary.LittleEndian.Uint32(header[8:12]),
		Compression:    compression,
		Metadata:       metadata,
		Payload:        payload,
	}, offset + headerLen + payloadLen, nil
}

func findFooter(data []byte) (int, bool) {
	if len(data) < 36 {
		return 0, false
	}
	offset := len(data) - 36
	return offset, bytes.Equal(data[offset:offset+4], footerMagic[:]) &&
		bytes.Equal(data[offset+32:offset+36], footerEndMagic[:])
}

func canonicalJSON(v map[string]any) ([]byte, error) {
	if v == nil {
		v = map[string]any{}
	}
	return json.Marshal(v)
}

func requireMetadata(metadata map[string]any, keys ...string) error {
	for _, key := range keys {
		if _, ok := metadata[key]; !ok {
			return fmt.Errorf("mrx: missing metadata %q", key)
		}
	}
	return nil
}
