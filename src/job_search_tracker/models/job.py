"""Normalized job model shared by search and application tracking."""

from __future__ import annotations

from dataclasses import dataclass

@dataclass(slots=True)
class Job:
    """Normalized job returned by source ingestion."""

    source: str
    external_id: str | None
    title: str
    company: str | None = None
    location: str | None = None
    url: str | None = None
    description: str | None = None
    skills: tuple[str, ...] = ()
    date_posted: str | None = None
    is_remote: bool | None = None
