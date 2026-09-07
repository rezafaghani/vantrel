import tempfile
import unittest
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
import check_questdb_schema


class QuestDBSchemaTest(unittest.TestCase):
    def test_valid_schema(self):
        self.assertEqual([], check_questdb_schema.check(Path("contracts/questdb/market_observations.sql")))

    def test_rejects_catalog_semantics(self):
        errors = self.check_text(
            """
CREATE TABLE IF NOT EXISTS market_actual_observations (
  series_id SYMBOL,
  event_time TIMESTAMP,
  value DOUBLE,
  unit SYMBOL,
  revision LONG,
  display_name STRING
) TIMESTAMP(event_time) PARTITION BY DAY WAL;
CREATE TABLE IF NOT EXISTS market_forecast_observations (
  series_id SYMBOL,
  forecast_run_id SYMBOL,
  issued_at TIMESTAMP,
  target_time TIMESTAMP,
  horizon_seconds LONG,
  value DOUBLE,
  unit SYMBOL
) TIMESTAMP(target_time) PARTITION BY DAY WAL;
"""
        )
        self.assertTrue(any("Catalog semantics" in error for error in errors), errors)

    def test_rejects_postgres_drift(self):
        errors = self.check_text("CREATE TABLE IF NOT EXISTS market_actual_observations (id SERIAL) TIMESTAMP(id);")
        self.assertTrue(any("non-QuestDB marker SERIAL" in error for error in errors), errors)

    def check_text(self, text: str) -> list[str]:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "schema.sql"
            path.write_text(text)
            return check_questdb_schema.check(path)


if __name__ == "__main__":
    unittest.main()
