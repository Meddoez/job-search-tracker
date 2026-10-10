"""Tests for environment-based database settings."""

import pytest

from job_search_tracker.settings import database_url


def test_database_url_rejects_missing_environment_value(monkeypatch):
    monkeypatch.delenv("DATABASE_URL", raising=False)

    with pytest.raises(ValueError, match="DATABASE_URL must be set"):
        database_url()


@pytest.mark.parametrize("value", ["", "   "])
def test_database_url_rejects_blank_environment_value(monkeypatch, value):
    monkeypatch.setenv("DATABASE_URL", value)

    with pytest.raises(ValueError, match="DATABASE_URL must be set"):
        database_url()


def test_database_url_uses_environment_value(monkeypatch):
    configured = "postgresql+psycopg://user:password@db:5432/application"
    monkeypatch.setenv("DATABASE_URL", configured)

    assert database_url() == configured
