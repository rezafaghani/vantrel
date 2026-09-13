package marketstore

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	actualTable   = "market_actual_observations"
	forecastTable = "market_forecast_observations"
	tradeTable    = "market_trade_observations"
	quoteTable    = "market_quote_observations"
)

type ILPWriter struct {
	w io.Writer
}

func NewILPWriter(w io.Writer) *ILPWriter {
	return &ILPWriter{w: w}
}

func (w *ILPWriter) WriteActual(obs ActualObservation) error {
	if w == nil || w.w == nil {
		return errors.New("marketstore ilp: nil writer")
	}
	if err := ValidateActual(obs); err != nil {
		return err
	}
	_, err := io.WriteString(w.w, actualLine(obs))
	return err
}

func (w *ILPWriter) WriteForecast(obs ForecastObservation) error {
	if w == nil || w.w == nil {
		return errors.New("marketstore ilp: nil writer")
	}
	if err := ValidateForecast(obs); err != nil {
		return err
	}
	_, err := io.WriteString(w.w, forecastLine(obs))
	return err
}

func (w *ILPWriter) WriteTrade(obs TradeObservation) error {
	if w == nil || w.w == nil {
		return errors.New("marketstore ilp: nil writer")
	}
	if err := ValidateTrade(obs); err != nil {
		return err
	}
	_, err := io.WriteString(w.w, tradeLine(obs))
	return err
}

func (w *ILPWriter) WriteQuote(obs QuoteObservation) error {
	if w == nil || w.w == nil {
		return errors.New("marketstore ilp: nil writer")
	}
	if err := ValidateQuote(obs); err != nil {
		return err
	}
	_, err := io.WriteString(w.w, quoteLine(obs))
	return err
}

func actualLine(obs ActualObservation) string {
	tags := tags(
		"series_id", obs.SeriesID,
		"unit", obs.Unit,
		"quality_state", obs.Quality.State,
		"raw_record_id", obs.Lineage.RawRecordID,
		"canonical_event_id", obs.Lineage.CanonicalEventID,
		"ingestion_id", obs.Lineage.IngestionID,
	)
	fields := fields(
		floatField("value", obs.Value),
		floatField("quality_score", obs.Quality.Score),
		intField("revision", int64(obs.Revision)),
		stringField("quality_flags", strings.Join(obs.Quality.Flags, ",")),
	)
	return line(actualTable, tags, fields, obs.EventTime.UnixNano())
}

func forecastLine(obs ForecastObservation) string {
	tags := tags(
		"series_id", obs.SeriesID,
		"forecast_run_id", obs.ForecastRunID,
		"unit", obs.Unit,
		"model_version", obs.ModelVersion,
		"quality_state", obs.Quality.State,
		"raw_record_id", obs.Lineage.RawRecordID,
		"canonical_event_id", obs.Lineage.CanonicalEventID,
		"ingestion_id", obs.Lineage.IngestionID,
	)
	fields := fields(
		timestampField("issued_at", obs.IssuedAt.UnixMicro()),
		intField("horizon_seconds", int64(obs.Horizon.Seconds())),
		floatField("value", obs.Value),
		floatField("quality_score", obs.Quality.Score),
		stringField("quality_flags", strings.Join(obs.Quality.Flags, ",")),
	)
	return line(forecastTable, tags, fields, obs.TargetTime.UnixNano())
}

func tradeLine(obs TradeObservation) string {
	tags := tags(
		"series_id", obs.SeriesID,
		"instrument", obs.Instrument,
		"currency", obs.Currency,
		"quality_state", obs.Quality.State,
		"raw_record_id", obs.Lineage.RawRecordID,
		"canonical_event_id", obs.Lineage.CanonicalEventID,
		"ingestion_id", obs.Lineage.IngestionID,
	)
	fields := fields(
		floatField("price", obs.Price),
		floatField("quantity", obs.Quantity),
		floatField("quality_score", obs.Quality.Score),
		stringField("quality_flags", strings.Join(obs.Quality.Flags, ",")),
	)
	return line(tradeTable, tags, fields, obs.TradeTime.UnixNano())
}

func quoteLine(obs QuoteObservation) string {
	tags := tags(
		"series_id", obs.SeriesID,
		"instrument", obs.Instrument,
		"currency", obs.Currency,
		"quality_state", obs.Quality.State,
		"raw_record_id", obs.Lineage.RawRecordID,
		"canonical_event_id", obs.Lineage.CanonicalEventID,
		"ingestion_id", obs.Lineage.IngestionID,
	)
	fields := fields(
		floatField("bid", obs.Bid),
		floatField("ask", obs.Ask),
		floatField("quantity", obs.Quantity),
		floatField("quality_score", obs.Quality.Score),
		stringField("quality_flags", strings.Join(obs.Quality.Flags, ",")),
	)
	return line(quoteTable, tags, fields, obs.QuoteTime.UnixNano())
}

func line(table string, tags, fields []string, timestamp int64) string {
	var b strings.Builder
	b.WriteString(table)
	for _, tag := range tags {
		b.WriteByte(',')
		b.WriteString(tag)
	}
	b.WriteByte(' ')
	b.WriteString(strings.Join(fields, ","))
	b.WriteByte(' ')
	b.WriteString(strconv.FormatInt(timestamp, 10))
	b.WriteByte('\n')
	return b.String()
}

func tags(kv ...string) []string {
	out := make([]string, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		if kv[i+1] != "" {
			out = append(out, escape(kv[i])+"="+escape(kv[i+1]))
		}
	}
	return out
}

func fields(values ...string) []string {
	out := values[:0]
	for _, value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func floatField(name string, value float64) string {
	return escape(name) + "=" + strconv.FormatFloat(value, 'f', -1, 64)
}

func intField(name string, value int64) string {
	return escape(name) + "=" + strconv.FormatInt(value, 10) + "i"
}

func timestampField(name string, micros int64) string {
	return escape(name) + "=" + strconv.FormatInt(micros, 10) + "t"
}

func stringField(name, value string) string {
	if value == "" {
		return ""
	}
	return escape(name) + `="` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}

func escape(value string) string {
	replacer := strings.NewReplacer(",", `\,`, " ", `\ `, "=", `\=`)
	return replacer.Replace(value)
}
