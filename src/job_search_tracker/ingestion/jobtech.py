"""Fetch and normalize jobs from JobTech."""

from __future__ import annotations

import re
from typing import Any

import requests

from job_search_tracker.extraction.matcher import extract_skills_from_texts
from job_search_tracker.models import Job
from job_search_tracker.normalization import required_string, string_or_none

API_URL = "https://jobsearch.api.jobtechdev.se/search"
DEFAULT_QUERY = "python"
PAGE_SIZE = 100
MAX_OFFSET = 2000
STRICT_SEARCH_HEADERS = {
    "x-feature-disable-smart-freetext": "true",
    "x-feature-freetext-bool-method": "and",
}
DEVELOPER_EQUIVALENTS = ("developer", "utvecklare", "systemutvecklare")


def fetch_jobs(query: str = DEFAULT_QUERY, limit: int | None = None, timeout: float = 30.0) -> list[Job]:
    """Fetch JobTech postings for a query and normalize them.

    Args:
        query: Search query sent to JobTech. Every comma- or whitespace-separated term is required.
        limit: Maximum number of postings to request from JobTech.
        timeout: HTTP request timeout in seconds.

    Returns:
        Normalized JobTech postings.

    Raises:
        ConnectionError: If the HTTP request fails.
        ValueError: If JobTech returns invalid JSON or an unexpected payload shape.
    """
    jobs: list[Job] = []
    seen_keys: set[tuple[str, str]] = set()
    required_term_groups = _required_term_groups(query)
    for query_variant in _query_variants(query):
        remaining = None if limit is None else limit - len(jobs)
        if remaining is not None and remaining <= 0:
            break

        for job in _fetch_jobs_for_query(
            query_variant,
            required_term_groups=required_term_groups,
            limit=remaining,
            timeout=timeout,
            excluded_keys=seen_keys,
        ):
            dedupe_key = _job_dedupe_key(job)
            if dedupe_key in seen_keys:
                continue
            seen_keys.add(dedupe_key)
            jobs.append(job)
            if limit is not None and len(jobs) >= limit:
                break

    return jobs[:limit] if limit is not None else jobs


def _fetch_jobs_for_query(
    query: str,
    required_term_groups: tuple[tuple[str, ...], ...],
    limit: int | None,
    timeout: float,
    excluded_keys: set[tuple[str, str]],
) -> list[Job]:
    if limit is None:
        payload = _fetch_json(query=query, limit=None, offset=None, timeout=timeout)
        return [
            job
            for job in _normalize_response_jobs(payload)
            if _job_matches_required_terms(job, required_term_groups)
        ]

    jobs: list[Job] = []
    seen_keys = set(excluded_keys)
    offset = 0
    while len(jobs) < limit and offset <= MAX_OFFSET:
        page_limit = min(PAGE_SIZE, limit - len(jobs))
        payload = _fetch_json(query=query, limit=page_limit, offset=offset, timeout=timeout)
        page_jobs = _normalize_response_jobs(payload)
        if not page_jobs:
            break

        for job in page_jobs:
            dedupe_key = _job_dedupe_key(job)
            if dedupe_key in seen_keys or not _job_matches_required_terms(job, required_term_groups):
                continue
            seen_keys.add(dedupe_key)
            jobs.append(job)
        if len(page_jobs) < page_limit:
            break
        offset += page_limit

    return jobs[:limit]


def _required_term_groups(query: str) -> tuple[tuple[str, ...], ...]:
    groups: list[tuple[str, ...]] = []
    seen: set[tuple[str, ...]] = set()
    for term in re.split(r"[\s,]+", query):
        normalized_term = term.strip()
        if not normalized_term:
            continue

        if normalized_term.casefold() in DEVELOPER_EQUIVALENTS:
            group = DEVELOPER_EQUIVALENTS
        else:
            group = (normalized_term,)
        group_key = tuple(value.casefold() for value in group)
        if group_key in seen:
            continue
        seen.add(group_key)
        groups.append(group)
    return tuple(groups)


def _job_matches_required_terms(
    job: Job,
    required_term_groups: tuple[tuple[str, ...], ...],
) -> bool:
    searchable_text = "\n".join(
        value for value in (job.title, job.description, job.company, job.location) if value
    )
    return all(
        any(_contains_word(searchable_text, term) for term in group)
        for group in required_term_groups
    )


def _query_variants(query: str) -> list[str]:
    normalized_query = " ".join(query.split())
    lower_query = normalized_query.lower()
    variants = [normalized_query]

    if _contains_word(lower_query, "developer"):
        variants.append(_replace_word(normalized_query, "developer", "utvecklare"))
        variants.append(_replace_word(normalized_query, "developer", "systemutvecklare"))
    if _contains_word(lower_query, "utvecklare"):
        variants.append(_replace_word(normalized_query, "utvecklare", "developer"))
        variants.append(_replace_word(normalized_query, "utvecklare", "systemutvecklare"))

    unique_variants: list[str] = []
    seen: set[str] = set()
    for variant in variants:
        normalized_variant = " ".join(variant.split())
        variant_key = normalized_variant.lower()
        if not normalized_variant or variant_key in seen:
            continue
        seen.add(variant_key)
        unique_variants.append(normalized_variant)
    return unique_variants


def _contains_word(text: str, word: str) -> bool:
    return re.search(rf"(?<!\w){re.escape(word)}(?!\w)", text, flags=re.IGNORECASE) is not None


def _replace_word(text: str, old: str, new: str) -> str:
    return re.sub(rf"(?<!\w){re.escape(old)}(?!\w)", new, text, flags=re.IGNORECASE)


def _job_dedupe_key(job: Job) -> tuple[str, str]:
    if job.external_id:
        return (job.source, job.external_id)
    return (job.source, "|".join((job.title, job.company or "", job.url or "")))


def normalize_jobtech_job(item: dict[str, Any]) -> Job:
    """Convert one JobTech API item into the internal Job model."""
    workplace_address = item.get("workplace_address") or {}
    employer = item.get("employer") or {}
    description = item.get("description") or {}
    skill_summary = _jobtech_skill_summary(item)
    description_text = string_or_none(description.get("text"))
    if skill_summary:
        description_text = skill_summary if not description_text else f"{skill_summary}\n\n{description_text}"
    skills = tuple(
        extract_skills_from_texts(
            (
                string_or_none(item.get("headline")) or "",
                description_text or "",
            )
        )
    )
    return Job(
        source="jobtech",
        external_id=string_or_none(item.get("id") or item.get("external_id") or item.get("original_id")),
        title=required_string(item.get("headline"), "headline"),
        company=string_or_none(employer.get("name") or employer.get("workplace")),
        location=string_or_none(
            workplace_address.get("city")
            or workplace_address.get("municipality")
            or workplace_address.get("region")
        ),
        url=string_or_none(item.get("webpage_url")),
        description=description_text,
        skills=skills,
        date_posted=string_or_none(item.get("publication_date")),
        is_remote=bool(item.get("remote")) if item.get("remote") is not None else None,
    )


def _fetch_json(query: str, limit: int | None, offset: int | None, timeout: float) -> dict[str, Any]:
    """Return the raw JobTech JSON search payload for a query."""
    params: dict[str, Any] = {"q": query}
    if limit is not None:
        params["limit"] = limit
    if offset is not None:
        params["offset"] = offset

    try:
        response = requests.get(
            API_URL,
            params=params,
            headers=STRICT_SEARCH_HEADERS,
            timeout=timeout,
        )
        response.raise_for_status()
    except requests.RequestException as exc:
        raise ConnectionError("Failed to fetch jobs from JobTech") from exc

    try:
        payload = response.json()
    except ValueError as exc:
        raise ValueError("JobTech response was not valid JSON") from exc

    if not isinstance(payload, dict):
        raise ValueError("JobTech response was not a JSON object")
    return payload


def _normalize_response_jobs(payload: dict[str, Any]) -> list[Job]:
    items = payload.get("hits")
    if not isinstance(items, list):
        items = payload.get("data")

    if not isinstance(items, list):
        raise ValueError("JobTech response did not include a job list")

    return [normalize_jobtech_job(item) for item in items if isinstance(item, dict)]


def _jobtech_skill_summary(item: dict[str, Any]) -> str | None:
    """Flatten JobTech skill fields into text for the shared skill matcher."""
    skill_labels = list(_iter_skill_labels((item.get("must_have") or {}).get("skills")))
    skill_labels.extend(_iter_skill_labels((item.get("nice_to_have") or {}).get("skills")))

    unique_labels: list[str] = []
    seen: set[str] = set()
    for label in skill_labels:
        normalized = label.strip()
        if not normalized or normalized in seen:
            continue
        seen.add(normalized)
        unique_labels.append(normalized)

    if not unique_labels:
        return None
    return "\n".join(unique_labels)


def _iter_skill_labels(value: Any) -> list[str]:
    """Return skill labels from nested JobTech skill values."""
    labels: list[str] = []
    if value is None:
        return labels

    if isinstance(value, str):
        text = string_or_none(value)
        if text:
            labels.append(text)
        return labels

    if isinstance(value, dict):
        for key in ("label", "name", "text"):
            text = string_or_none(value.get(key))
            if text:
                labels.append(text)
                break
        else:
            for nested_value in value.values():
                labels.extend(_iter_skill_labels(nested_value))
        return labels

    if isinstance(value, list):
        for entry in value:
            labels.extend(_iter_skill_labels(entry))

    return labels
