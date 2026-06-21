"""File-backed TOFU KnownKeyStore for blerpc E2E identity pinning.

The TOFU policy (pin on first use, reject a changed key) now lives in the
``blerpc_protocol`` library (``tofu_verify`` + ``KnownKeyStore``); this module
only provides the platform persistence — a JSON file — as a ``KnownKeyStore``.
"""

from __future__ import annotations

import json
import logging
import os

logger = logging.getLogger(__name__)


class FileKnownKeyStore:
    """A ``blerpc_protocol.KnownKeyStore`` backed by a JSON file (per-user)."""

    def __init__(self, path: str) -> None:
        self._path = path

    def get(self, device_id: str) -> str | None:
        return _load_known_keys(self._path).get(device_id)

    def put(self, device_id: str, hex_ed25519_pubkey: str) -> None:
        known = _load_known_keys(self._path)
        known[device_id] = hex_ed25519_pubkey
        _save_known_keys(self._path, known)
        logger.info("Stored key for %s (TOFU)", device_id)


def _load_known_keys(path: str) -> dict[str, str]:
    """Load known keys from JSON file."""
    if not os.path.exists(path):
        return {}
    try:
        with open(path) as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError):
        return {}


def _save_known_keys(path: str, known: dict[str, str]) -> None:
    """Save known keys to JSON file with restricted permissions (0600)."""
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, "w") as f:
        json.dump(known, f, indent=2)
