from __future__ import annotations

import importlib.util
import json
import pathlib
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
SPEC = importlib.util.spec_from_file_location(
    "gen_authored_bios", ROOT / "tools" / "gen_authored_bios.py"
)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class AuthoredBiosTest(unittest.TestCase):
    def write(self, text: str) -> pathlib.Path:
        tmp = tempfile.NamedTemporaryFile("w", suffix=".md", encoding="utf-8", delete=False)
        self.addCleanup(pathlib.Path(tmp.name).unlink, missing_ok=True)
        with tmp:
            tmp.write(text)
        return pathlib.Path(tmp.name)

    def test_current_batches_are_complete_and_deterministic(self) -> None:
        first, summary = MODULE.build_document()
        second, summary2 = MODULE.build_document()
        self.assertEqual(summary, summary2)
        self.assertEqual(417, summary["facts"])
        self.assertEqual(387, summary["authored"])
        self.assertEqual(30, summary["unknown"])
        self.assertEqual(61, len(summary["newly_available"]))
        self.assertEqual(
            json.dumps(first, ensure_ascii=False, indent=2) + "\n",
            json.dumps(second, ensure_ascii=False, indent=2) + "\n",
        )

    def test_rejects_malformed_person_header(self) -> None:
        path = self.write("## #1 蔣中正 medium\n\n正文。\n")
        with self.assertRaisesRegex(MODULE.InputError, "無法解析人物標題"):
            MODULE.parse_sections(path)

    def test_unknown_requires_explicit_no_biography_marker(self) -> None:
        path = self.write("## #8 朱俊光 — `unknown`\n\n只有查詢紀錄。\n")
        with self.assertRaisesRegex(MODULE.InputError, "沒有明示不立傳"):
            MODULE.parse_sections(path)

    def test_unknown_audit_text_is_not_mistaken_for_biography(self) -> None:
        path = self.write(
            "## #8 朱俊光 — `unknown`（不寫小傳）\n\n"
            "查無可用記載，bio_zh` 留空。\n\n- **查證範圍**：零命中\n"
        )
        section = MODULE.parse_sections(path)[8]
        self.assertEqual("", section["bio_zh"])


if __name__ == "__main__":
    unittest.main()
