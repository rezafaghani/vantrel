package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rezafaghani/vantrel/libs/go/ingestion"
	"github.com/rezafaghani/vantrel/libs/go/marketstore"
	"github.com/rezafaghani/vantrel/libs/go/rawarchive"
	"github.com/rezafaghani/vantrel/libs/go/synthetic"
	"github.com/rezafaghani/vantrel/libs/go/validation"
)

func main() {
	steps := flag.Int("steps", 3, "synthetic market steps")
	flag.Parse()
	if err := run(context.Background(), *steps, time.Now().UTC(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, steps int, now time.Time, out io.Writer) error {
	if steps <= 0 {
		return errors.New("steps must be positive")
	}
	adapter, err := synthetic.Adapter(synthetic.Config{Seed: 1, Steps: steps})
	if err != nil {
		return err
	}
	var raw bytes.Buffer
	sink, err := rawarchive.NewSink(&raw, synthetic.ProviderID, synthetic.SourceID, now)
	if err != nil {
		return err
	}
	var ilp bytes.Buffer
	err = ingestion.Runtime{
		Adapter: adapter,
		Mode:    ingestion.ModeHistorical,
		Hooks: ingestion.Hooks{
			RawSink:   sink,
			Parser:    synthetic.Parser{},
			Validator: validation.Engine{},
			Publisher: marketstore.Publisher{Store: marketstore.NewILPWriter(&ilp)},
		},
	}.Run(ctx)
	if closeErr := sink.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "vantrel local pipeline\nsteps=%d\nraw_archive_bytes=%d\nmarket_ilp_lines=%d\nfirst_ilp_line=%s\n", steps, raw.Len(), strings.Count(ilp.String(), "\n"), firstLine(ilp.String()))
	return err
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
