"""Representative, synthetic JobTech responses shared by service tests."""

import json
from pathlib import Path

import pytest


@pytest.fixture
def jobtech_search_payload():
    fixture_path = Path(__file__).parent / "fixtures" / "jobtech_search.json"
    return json.loads(fixture_path.read_text(encoding="utf-8"))
