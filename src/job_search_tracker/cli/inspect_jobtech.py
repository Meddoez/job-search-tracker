"""Inspect and save one JobTech JobSearch API response."""

from __future__ import annotations

import json
from pathlib import Path

import requests


API_URL = "https://jobsearch.api.jobtechdev.se/search"
DEFAULT_QUERY = "python"


def fetch_jobtech_sample(query: str = DEFAULT_QUERY, timeout: float = 30.0) -> dict:
    """Fetch one JobTech search response for inspection."""
    response = requests.get(API_URL, params={"q": query}, timeout=timeout)
    print(f"HTTP status: {response.status_code}")
    response.raise_for_status()
    payload = response.json()
    if not isinstance(payload, dict):
        raise ValueError("JobTech response was not a JSON object")
    return payload


def result_count(payload: dict) -> int:
    """Return the number of results in the JobTech response."""
    for key in ("hits", "results", "data", "items"):
        value = payload.get(key)
        if isinstance(value, list):
            return len(value)
    return 0


def save_raw_response(payload: dict, output_path: Path) -> None:
    """Persist the raw JSON response for local inspection."""
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(payload, indent=2, sort_keys=True), encoding="utf-8")


def main() -> None:
    payload = fetch_jobtech_sample()
    print(f"Results returned: {result_count(payload)}")
    repo_root = Path(__file__).resolve().parents[3]
    output_path = repo_root / "data/raw/jobtech_sample.json"
    save_raw_response(payload, output_path)
    print(f"Saved raw response to: {output_path}")


if __name__ == "__main__":
    main()
