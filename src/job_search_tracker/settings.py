"""Application settings loaded from environment variables."""

from __future__ import annotations

import os


def database_url() -> str:
    """Return the configured PostgreSQL database URL."""
    url = os.environ.get("DATABASE_URL")
    if not url or not url.strip():
        raise ValueError("DATABASE_URL must be set")
    return url
