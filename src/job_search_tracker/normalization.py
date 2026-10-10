"""Shared job normalization helpers."""

from __future__ import annotations

from typing import Any


def string_or_none(value: Any) -> str | None:
    if value is None:
        return None
    text = str(value).strip()
    return text if text else None


def required_string(value: Any, field_name: str) -> str:
    text = string_or_none(value)
    if text is None:
        raise ValueError(f"{field_name} is required")
    return text
