CREATE TABLE IF NOT EXISTS market_actual_observations (
  series_id SYMBOL CAPACITY 1048576 CACHE,
  event_time TIMESTAMP,
  value DOUBLE,
  unit SYMBOL CAPACITY 1024 CACHE,
  quality_state SYMBOL CAPACITY 256 CACHE,
  quality_score DOUBLE,
  quality_flags STRING,
  revision LONG,
  raw_record_id SYMBOL CAPACITY 1048576 NOCACHE,
  canonical_event_id SYMBOL CAPACITY 1048576 NOCACHE,
  ingestion_id SYMBOL CAPACITY 1048576 NOCACHE
) TIMESTAMP(event_time) PARTITION BY DAY WAL;

CREATE TABLE IF NOT EXISTS market_forecast_observations (
  series_id SYMBOL CAPACITY 1048576 CACHE,
  forecast_run_id SYMBOL CAPACITY 1048576 NOCACHE,
  issued_at TIMESTAMP,
  target_time TIMESTAMP,
  horizon_seconds LONG,
  value DOUBLE,
  unit SYMBOL CAPACITY 1024 CACHE,
  model_version SYMBOL CAPACITY 1048576 NOCACHE,
  quality_state SYMBOL CAPACITY 256 CACHE,
  quality_score DOUBLE,
  quality_flags STRING,
  raw_record_id SYMBOL CAPACITY 1048576 NOCACHE,
  canonical_event_id SYMBOL CAPACITY 1048576 NOCACHE,
  ingestion_id SYMBOL CAPACITY 1048576 NOCACHE
) TIMESTAMP(target_time) PARTITION BY DAY WAL;
