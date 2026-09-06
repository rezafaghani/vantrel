import json
import tempfile
import unittest
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
import check_kafka_topics


class KafkaTopicsTest(unittest.TestCase):
    def check_data(self, data: dict) -> list[str]:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "topics.json"
            path.write_text(json.dumps(data))
            return check_kafka_topics.check(path)

    def test_valid_registry(self):
        self.assertEqual([], check_kafka_topics.check(Path("contracts/kafka/topics.json")))

    def test_rejects_duplicate_topic(self):
        errors = self.check_data(
            {
                "version": 1,
                "topics": [
                    topic("vantrel.market.canonical.v1"),
                    topic("vantrel.market.canonical.v1"),
                ],
            }
        )
        self.assertTrue(any("duplicate topic name" in error for error in errors), errors)

    def test_rejects_unversioned_topic(self):
        errors = self.check_data({"version": 1, "topics": [topic("vantrel.market.canonical")]})
        self.assertTrue(any("invalid topic name" in error for error in errors), errors)


def topic(name: str) -> dict:
    return {
        "name": name,
        "owner": "test",
        "purpose": "test stream",
        "key": "series_id",
        "value_contract": "test contract",
        "partitions": 1,
        "replication_factor": 1,
        "retention_ms": 1,
        "cleanup_policy": "delete",
        "replayable": True,
    }


if __name__ == "__main__":
    unittest.main()
