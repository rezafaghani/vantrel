package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/mrx"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stderr)
		return errors.New("missing command")
	}
	switch args[0] {
	case "inspect":
		return inspect(args[1:], stdout)
	case "verify":
		return verify(args[1:], stdout)
	case "extract":
		return extract(args[1:], stdout)
	case "benchmark":
		return benchmark(args[1:], stdout)
	case "-h", "--help", "help":
		usage(stdout)
		return nil
	default:
		usage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func inspect(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: mrx inspect <archive.mrx>")
	}
	a, err := readArchive(fs.Arg(0))
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(map[string]any{
		"status":        a.Status,
		"frames":        len(a.Frames),
		"created_at_ns": a.Header.CreatedAtNS,
		"metadata":      a.Header.Metadata,
		"index":         a.Index,
	})
}

func verify(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: mrx verify <archive.mrx>")
	}
	a, err := readArchive(fs.Arg(0))
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s frames=%d\n", a.Status, len(a.Frames))
	if a.Status != mrx.StatusClean {
		return a.Err
	}
	return nil
}

func extract(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	frame := fs.Uint64("frame", 0, "frame number")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: mrx extract [-frame N] <archive.mrx>")
	}
	a, err := readArchive(fs.Arg(0))
	if err != nil {
		return err
	}
	f, ok := a.FrameByNumber(*frame)
	if !ok {
		return fmt.Errorf("frame %d not found", *frame)
	}
	_, err = stdout.Write(f.Payload)
	return err
}

func benchmark(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	frames := fs.Int("frames", 1000, "number of frames")
	size := fs.Int("size", 256, "payload bytes per frame")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *frames < 1 || *size < 1 {
		return errors.New("frames and size must be positive")
	}
	payload := make([]byte, *size)
	for i := range payload {
		payload[i] = byte(i)
	}
	var sink countingWriter
	start := time.Now()
	w, err := mrx.NewWriter(&sink, mrx.Header{
		CreatedAtNS: time.Now().UnixNano(),
		Metadata: map[string]any{
			"creator":     "mrx-cli",
			"provider_id": "benchmark",
			"source_id":   "benchmark",
		},
	})
	if err != nil {
		return err
	}
	for i := 0; i < *frames; i++ {
		if err := w.WriteFrame(mrx.Frame{
			ReceiveTimeNS:  time.Now().UnixNano(),
			SourceTimeNS:   -1,
			SourceSequence: int64(i),
			Compression:    mrx.CompressionNone,
			Metadata: map[string]any{
				"encoding":    "binary",
				"provider_id": "benchmark",
				"source_id":   "benchmark",
			},
			Payload: payload,
		}); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	elapsed := time.Since(start)
	fmt.Fprintf(stdout, "frames=%d payload_bytes=%d archive_bytes=%d elapsed=%s\n", *frames, (*frames)*(*size), sink.n, elapsed)
	return nil
}

func readArchive(path string) (mrx.Archive, error) {
	f, err := os.Open(path)
	if err != nil {
		return mrx.Archive{}, err
	}
	defer f.Close()
	a := mrx.Read(f)
	if a.Err != nil && a.Status != mrx.StatusUncleanRecoverable {
		return a, a.Err
	}
	return a, nil
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: mrx <inspect|verify|extract|benchmark> [options]")
}

type countingWriter struct {
	n int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	return len(p), nil
}
