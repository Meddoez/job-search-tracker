"""Shared test helpers."""

from __future__ import annotations

import requests


class DummyResponse:
    def __init__(self, payload, status_code=200, json_error=None):
        self._payload = payload
        self.status_code = status_code
        self._json_error = json_error

    def raise_for_status(self):
        if self.status_code >= 400:
            raise requests.HTTPError(f"{self.status_code} error")

    def json(self):
        if self._json_error is not None:
            raise self._json_error
        return self._payload
