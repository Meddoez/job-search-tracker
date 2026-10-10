"""Tests for JobTech ingestion and normalization."""

import pytest
import requests

from job_search_tracker.ingestion import normalize_jobtech_job
from job_search_tracker.ingestion.jobtech import fetch_jobs
from tests.helpers import DummyResponse


STRICT_SEARCH_HEADERS = {
    "x-feature-disable-smart-freetext": "true",
    "x-feature-freetext-bool-method": "and",
}


def jobtech_item(job_id, *, headline=None, description="Build Python services."):
    return {
        "id": str(job_id),
        "headline": headline or f"Python Developer {job_id}",
        "employer": {"name": "Example GmbH"},
        "workplace_address": {"city": "Stockholm"},
        "webpage_url": f"https://jobtech.example/jobs/{job_id}",
        "description": {"text": description},
        "publication_date": "2026-03-16T13:53:01",
        "remote": True,
    }


def test_normalize_jobtech_job_maps_expected_fields():
    raw = {
        "id": "30751757",
        "headline": "Python Developer",
        "employer": {
            "name": "Example GmbH",
            "workplace": "Example Workplace",
        },
        "workplace_address": {
            "city": "Stockholm",
            "municipality": "Stockholm",
            "region": "Stockholms län",
        },
        "webpage_url": "https://jobtech.example/jobs/example",
        "description": {
            "text": "Build Python services.",
        },
        "publication_date": "2026-03-16T13:53:01",
        "remote": True,
        "tags": ["Python", "Backend"],
    }

    job = normalize_jobtech_job(raw)

    assert job.source == "jobtech"
    assert job.external_id == "30751757"
    assert job.title == "Python Developer"
    assert job.company == "Example GmbH"
    assert job.location == "Stockholm"
    assert job.url == "https://jobtech.example/jobs/example"
    assert job.description == "Build Python services."
    assert job.date_posted == "2026-03-16T13:53:01"
    assert job.is_remote is True


def test_normalize_jobtech_job_handles_missing_optional_fields():
    raw = {"headline": "Backend Engineer"}

    job = normalize_jobtech_job(raw)

    assert job.source == "jobtech"
    assert job.external_id is None
    assert job.title == "Backend Engineer"
    assert job.company is None
    assert job.location is None
    assert job.url is None
    assert job.description is None
    assert job.date_posted is None
    assert job.is_remote is None


def test_normalize_jobtech_job_preserves_date_posted_as_text():
    raw = {
        "id": "data-engineer-789",
        "headline": "Data Engineer",
        "publication_date": "not-a-date",
    }

    job = normalize_jobtech_job(raw)

    assert job.date_posted == "not-a-date"


def test_normalize_jobtech_job_handles_live_nested_fields():
    raw = {
        "id": "live-shaped-job",
        "headline": "Cloud Developer",
        "workplace_address": {"city": None, "municipality": "Stockholm"},
        "description": {"text": "Build services with Docker."},
        "must_have": {"skills": [{"label": "AWS"}]},
        "nice_to_have": {"skills": [{"name": "PostgreSQL"}]},
        "publication_date": "2026-07-15T08:12:55",
    }

    job = normalize_jobtech_job(raw)

    assert job.location == "Stockholm"
    assert job.date_posted == "2026-07-15T08:12:55"
    assert job.is_remote is None
    assert job.skills == ("docker", "postgresql", "cloud", "aws")


def test_fetch_jobs_normalizes_each_item(monkeypatch):
    payload = {
        "data": [
            {
                "id": "30751757",
                "headline": "Python Developer",
                "employer": {"name": "Example GmbH"},
                "workplace_address": {"city": "Stockholm"},
                "webpage_url": "https://jobtech.example/jobs/python-dev",
                "description": {"text": "Build Python services."},
                "publication_date": "2026-03-16T13:53:01",
                "remote": True,
            },
            "ignore-me",
        ]
    }

    def fake_get(url, params, headers, timeout):
        assert url == "https://jobsearch.api.jobtechdev.se/search"
        assert params == {"q": "python"}
        assert headers == STRICT_SEARCH_HEADERS
        assert timeout == 12.5
        return DummyResponse(payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(timeout=12.5)

    assert len(jobs) == 1
    assert jobs[0].external_id == "30751757"
    assert jobs[0].title == "Python Developer"
    assert jobs[0].company == "Example GmbH"
    assert jobs[0].location == "Stockholm"
    assert jobs[0].url == "https://jobtech.example/jobs/python-dev"
    assert jobs[0].description == "Build Python services."
    assert jobs[0].date_posted == "2026-03-16T13:53:01"
    assert jobs[0].is_remote is True


def test_fetch_jobs_sends_limit_to_jobtech(monkeypatch):
    payload = {"data": [jobtech_item(1)]}

    def fake_get(url, params, headers, timeout):
        assert url == "https://jobsearch.api.jobtechdev.se/search"
        assert params == {"q": "python", "limit": 25, "offset": 0}
        assert headers == STRICT_SEARCH_HEADERS
        assert timeout == 12.5
        return DummyResponse(payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    assert len(fetch_jobs(limit=25, timeout=12.5)) == 1


def test_fetch_jobs_paginates_until_limit(monkeypatch):
    calls = []

    def fake_get(url, params, headers, timeout):
        calls.append(params)
        start = params["offset"]
        page_size = params["limit"]
        payload = {"data": [jobtech_item(index) for index in range(start, start + page_size)]}
        return DummyResponse(payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(limit=250)

    assert len(jobs) == 250
    assert calls == [
        {"q": "python", "limit": 100, "offset": 0},
        {"q": "python", "limit": 100, "offset": 100},
        {"q": "python", "limit": 50, "offset": 200},
    ]


def test_fetch_jobs_expands_developer_query_and_deduplicates(monkeypatch):
    calls = []
    payloads = {
        "junior developer": {"data": [jobtech_item(1, headline="Junior Developer")]},
        "junior utvecklare": {
            "data": [
                jobtech_item(1, headline="Junior Developer"),
                jobtech_item(2, headline="Junior utvecklare"),
            ]
        },
        "junior systemutvecklare": {
            "data": [jobtech_item(3, headline="Junior systemutvecklare")]
        },
    }

    def fake_get(url, params, headers, timeout):
        calls.append(params)
        return DummyResponse(payloads[params["q"]])

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="junior developer", limit=10)

    assert [job.external_id for job in jobs] == ["1", "2", "3"]
    assert {call["q"] for call in calls} == set(payloads)
    assert len(jobs) <= 10


def test_fetch_jobs_does_not_expand_compound_utvecklare_word(monkeypatch):
    calls = []

    def fake_get(url, params, headers, timeout):
        calls.append(params)
        return DummyResponse({"data": [jobtech_item(1)]})

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="systemutvecklare", limit=10)

    assert [job.external_id for job in jobs] == ["1"]
    assert calls == [{"q": "systemutvecklare", "limit": 10, "offset": 0}]


def test_fetch_jobs_stops_when_page_is_short(monkeypatch):
    calls = []

    def fake_get(url, params, headers, timeout):
        calls.append(params)
        payload = {"data": [jobtech_item(1)]}
        return DummyResponse(payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(limit=100)

    assert len(jobs) == 1
    assert calls == [{"q": "python", "limit": 100, "offset": 0}]


def test_fetch_jobs_requires_every_query_term(monkeypatch):
    calls = []

    def fake_get(url, params, headers, timeout):
        calls.append(params)
        assert headers == STRICT_SEARCH_HEADERS
        if params["offset"] > 0:
            return DummyResponse({"data": []})
        return DummyResponse(
            {
                "data": [
                    jobtech_item(1, headline="Junior Developer", description="Build c++ services."),
                    jobtech_item(2, headline="Junior Developer", description="Build Go services."),
                    jobtech_item(3, headline="Senior C++ Developer"),
                ]
            }
        )

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="C++, Junior", limit=3)

    assert [job.external_id for job in jobs] == ["1"]
    assert calls == [
        {"q": "C++, Junior", "limit": 3, "offset": 0},
        {"q": "C++, Junior", "limit": 2, "offset": 3},
    ]


def test_fetch_jobs_accepts_swedish_developer_equivalent(monkeypatch):
    def fake_get(url, params, headers, timeout):
        if params["q"] == "utvecklare, C++, Junior":
            return DummyResponse(
                {
                    "data": [
                        jobtech_item(
                            1,
                            headline="Junior utvecklare",
                            description="Arbeta med C++.",
                        )
                    ]
                }
            )
        return DummyResponse({"data": []})

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="developer, C++, Junior", limit=1)

    assert [job.external_id for job in jobs] == ["1"]


def test_fetch_jobs_paginates_past_nonmatching_jobs(monkeypatch):
    calls = []

    def fake_get(url, params, headers, timeout):
        calls.append(params)
        if params["offset"] == 0:
            return DummyResponse(
                {
                    "data": [
                        jobtech_item(1, headline="Senior Python Developer"),
                        jobtech_item(2, headline="Senior Python Engineer"),
                    ]
                }
            )
        return DummyResponse(
            {
                "data": [
                    jobtech_item(3, headline="Junior Python Developer"),
                    jobtech_item(4, headline="Junior Python Engineer"),
                ]
            }
        )

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="python junior", limit=2)

    assert [job.external_id for job in jobs] == ["3", "4"]
    assert calls == [
        {"q": "python junior", "limit": 2, "offset": 0},
        {"q": "python junior", "limit": 2, "offset": 2},
    ]


@pytest.mark.parametrize("request_error", [requests.ConnectionError, requests.Timeout])
def test_fetch_jobs_raises_connection_error_on_request_failure(monkeypatch, request_error):
    upstream_error = request_error("JobTech unavailable")

    def fake_get(url, params, headers, timeout):
        raise upstream_error

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    with pytest.raises(ConnectionError, match="Failed to fetch jobs from JobTech") as exc_info:
        fetch_jobs()
    assert exc_info.value.__cause__ is upstream_error


def test_fetch_jobs_reads_hits_and_extracts_nested_skills(monkeypatch, jobtech_search_payload):
    def fake_get(url, params, headers, timeout):
        return DummyResponse(jobtech_search_payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="backend", limit=10)

    assert [job.external_id for job in jobs] == ["fixture-python-1", "fixture-java-2"]
    assert jobs[0].company == "Example AB"
    assert jobs[0].location == "Stockholm"
    assert jobs[0].skills == ("python", "docker", "postgresql", "aws")
    assert jobs[1].skills == ("java", "spring")


def test_fetch_jobs_applies_limit_to_unique_jobs_across_query_variants(monkeypatch):
    def fake_get(url, params, headers, timeout):
        if params["q"] == "junior developer":
            items = [jobtech_item(1, headline="Junior Developer")]
        else:
            items = [
                jobtech_item(1, headline="Junior Developer"),
                jobtech_item(2, headline="Junior utvecklare"),
                jobtech_item(3, headline="Junior systemutvecklare"),
            ]
        start = params["offset"]
        return DummyResponse({"hits": items[start:start + params["limit"]]})

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    jobs = fetch_jobs(query="junior developer", limit=2)

    assert [job.external_id for job in jobs] == ["1", "2"]


def test_fetch_jobs_paginates_past_duplicate_jobs(monkeypatch):
    items = [jobtech_item(1), jobtech_item(1), jobtech_item(2)]

    def fake_get(url, params, headers, timeout):
        start = params["offset"]
        return DummyResponse({"hits": items[start:start + params["limit"]]})

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    assert [job.external_id for job in fetch_jobs(query="python", limit=2)] == ["1", "2"]


@pytest.mark.parametrize("status_code", [429, 500, 503])
def test_fetch_jobs_raises_connection_error_on_http_error(monkeypatch, status_code):
    def fake_get(url, params, headers, timeout):
        return DummyResponse({"error": "unavailable"}, status_code=status_code)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    with pytest.raises(ConnectionError, match="Failed to fetch jobs from JobTech"):
        fetch_jobs()


def test_fetch_jobs_rejects_invalid_json(monkeypatch):
    def fake_get(url, params, headers, timeout):
        return DummyResponse(None, json_error=ValueError("invalid JSON"))

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    with pytest.raises(ValueError, match="JobTech response was not valid JSON"):
        fetch_jobs()


@pytest.mark.parametrize("payload", [None, [], "not an object"])
def test_fetch_jobs_rejects_non_object_json(monkeypatch, payload):
    def fake_get(url, params, headers, timeout):
        return DummyResponse(payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    with pytest.raises(ValueError, match="JobTech response was not a JSON object"):
        fetch_jobs()


@pytest.mark.parametrize("payload", [{}, {"total": {"value": 0}}, {"hits": {}}, {"data": "invalid"}])
def test_fetch_jobs_rejects_missing_or_invalid_job_list(monkeypatch, payload):
    def fake_get(url, params, headers, timeout):
        return DummyResponse(payload)

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    with pytest.raises(ValueError, match="JobTech response did not include a job list"):
        fetch_jobs()


@pytest.mark.parametrize("field", ["hits", "data"])
def test_fetch_jobs_accepts_empty_job_list(monkeypatch, field):
    def fake_get(url, params, headers, timeout):
        return DummyResponse({field: []})

    monkeypatch.setattr("job_search_tracker.ingestion.jobtech.requests.get", fake_get)

    assert fetch_jobs(limit=10) == []
