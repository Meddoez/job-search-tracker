"""Test configuration for local imports.

This keeps pytest runnable from the repo root without requiring PYTHONPATH.
"""

from __future__ import annotations

import os
import sys
from pathlib import Path

import pytest
from sqlalchemy import create_engine, delete
from sqlalchemy.engine import make_url


SRC_DIR = Path(__file__).resolve().parent / "src"
if str(SRC_DIR) not in sys.path:
    sys.path.insert(0, str(SRC_DIR))


@pytest.fixture
def postgres_url():
    """Return the disposable PostgreSQL URL or skip integration tests."""
    url = os.environ.get("TEST_DATABASE_URL")
    if not url:
        pytest.skip("TEST_DATABASE_URL is not configured")
    database_name = make_url(url).database
    if not database_name or not database_name.endswith("_test"):
        pytest.fail("TEST_DATABASE_URL must target a database whose name ends with '_test'")
    return url


@pytest.fixture
def clean_database(postgres_url):
    """Clear application tables before and after a PostgreSQL test."""
    from job_search_tracker.storage.metadata import (
        jobs,
        job_skills,
        profiles,
        profile_jobs,
        profile_skills,
    )

    engine = create_engine(postgres_url)

    def clear_tables():
        with engine.begin() as connection:
            connection.execute(delete(profile_jobs))
            connection.execute(delete(job_skills))
            connection.execute(delete(jobs))
            connection.execute(delete(profile_skills))
            connection.execute(delete(profiles))

    clear_tables()
    try:
        yield postgres_url
    finally:
        clear_tables()
        engine.dispose()
