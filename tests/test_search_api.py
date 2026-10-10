"""Tests for the internal job search HTTP API."""

import pytest
from fastapi.testclient import TestClient

from job_search_tracker.models import Job
from job_search_tracker.search_api import app
from tests.helpers import DummyResponse


client = TestClient(app, raise_server_exceptions=False)


def test_health() -> None:
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_search_returns_normalized_jobs(monkeypatch: pytest.MonkeyPatch) -> None:
    def fake_fetch_jobs(query: str, limit: int) -> list[Job]:
        assert query == "C++"
        assert limit == 50
        return [
            Job(
                source="jobtech",
                external_id="job-1",
                title="C++ Developer",
                company="Example AB",
                location=None,
                url="https://example.com/jobs/1",
                description="Build embedded software.",
                skills=("c++", "git"),
                date_posted="2026-09-26T10:00:00",
                is_remote=False,
            )
        ]

    monkeypatch.setattr("job_search_tracker.search_api.fetch_jobs", fake_fetch_jobs)

    response = client.post("/search", json={"query": "  C++  ", "limit": 50})

    assert response.status_code == 200
    assert response.json() == {
        "count": 1,
        "jobs": [
            {
                "source": "jobtech",
                "external_id": "job-1",
                "title": "C++ Developer",
                "company": "Example AB",
                "location": None,
                "url": "https://example.com/jobs/1",
                "description": "Build embedded software.",
                "skills": ["c++", "git"],
                "date_posted": "2026-09-26T10:00:00",
                "is_remote": False,
            }
        ],
    }


@pytest.mark.parametrize(
    "payload",
    [
        {"query": "", "limit": 1},
        {"query": "   ", "limit": 1},
        {"limit": 1},
        {"query": "python", "limit": 0},
        {"query": "python", "limit": 101},
    ],
)
def test_search_rejects_invalid_request(payload: dict[str, object]) -> None:
    response = client.post("/search", json=payload)

    assert response.status_code == 422


def test_search_maps_jobtech_connection_failure(monkeypatch: pytest.MonkeyPatch) -> None:
    def fail_fetch_jobs(query: str, limit: int) -> list[Job]:
        raise ConnectionError("JobTech unavailable")

    monkeypatch.setattr("job_search_tracker.search_api.fetch_jobs", fail_fetch_jobs)

    response = client.post("/search", json={"query": "python", "limit": 10})

    assert response.status_code == 502
    assert response.json() == {"detail": "JobTech request failed"}


def test_search_maps_invalid_jobtech_response(monkeypatch: pytest.MonkeyPatch) -> None:
    def fail_fetch_jobs(query: str, limit: int) -> list[Job]:
        raise ValueError("unexpected payload")

    monkeypatch.setattr("job_search_tracker.search_api.fetch_jobs", fail_fetch_jobs)

    response = client.post("/search", json={"query": "python", "limit": 10})

    assert response.status_code == 502
    assert response.json() == {"detail": "JobTech returned an invalid response"}


def test_search_normalizes_and_filters_jobtech_response(monkeypatch, jobtech_search_payload):
    def fake_get(url, params, headers, timeout):
        assert params["q"] == "python"
        return DummyResponse(jobtech_search_payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    response = client.post("/search", json={"query": "  python  "})

    assert response.status_code == 200
    assert response.json() == {
        "count": 1,
        "jobs": [
            {
                "source": "jobtech",
                "external_id": "fixture-python-1",
                "title": "Python Backend Engineer",
                "company": "Example AB",
                "location": "Stockholm",
                "url": "https://example.com/jobs/python-backend",
                "description": "Amazon Web Services (AWS)\nPostgreSQL\n\nBuild Python services with Docker.",
                "skills": ["python", "docker", "postgresql", "aws"],
                "date_posted": "2026-09-26T10:00:00",
                "is_remote": None,
            }
        ],
    }


def test_search_maps_malformed_upstream_payload_to_bad_gateway(monkeypatch):
    def fake_get(url, params, headers, timeout):
        return DummyResponse({"unexpected": "payload"})

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    response = client.post("/search", json={"query": "python"})

    assert response.status_code == 502
    assert response.json() == {"detail": "JobTech returned an invalid response"}
