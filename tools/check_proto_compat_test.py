import tempfile
import unittest
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
import check_proto_compat


class ProtoCompatTest(unittest.TestCase):
    def check_text(self, text: str) -> list[str]:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "test.proto"
            path.write_text(text)
            return check_proto_compat.check(path)

    def test_valid_proto(self):
        self.assertEqual(
            [],
            self.check_text(
                """
message Trade {
  string id = 1;
  reserved 100 to 199;
}
"""
            ),
        )

    def test_duplicate_field_number(self):
        errors = self.check_text(
            """
message Trade {
  string id = 1;
  string other_id = 1;
}
"""
        )
        self.assertIn("duplicate field 1", errors[0])

    def test_reserved_field_number(self):
        errors = self.check_text(
            """
message Trade {
  reserved 100 to 199;
  string future = 100;
}
"""
        )
        self.assertIn("field 100 is reserved", errors[0])


if __name__ == "__main__":
    unittest.main()
