"""TOFU (Trust On First Use) key management for blerpc E2E encryption."""

from __future__ import annotations

import json
import logging
import os

logger = logging.getLogger(__name__)


def check_or_store_key(
    known_keys_path: str, device_address: str, ed25519_pubkey: bytes
) -> bool:
    """Check a peripheral's Ed25519 public key against known keys.

    On first use, stores the key. On subsequent connections, verifies it matches.

    Returns True if the key is trusted (first use or matches stored key).
    Returns False if the key has changed (TOFU violation).
    """
    pubkey_hex = ed25519_pubkey.hex()
    known = _load_known_keys(known_keys_path)

    if device_address in known:
        stored_hex = known[device_address]
        if stored_hex == pubkey_hex:
            logger.info("Known key verified for %s", device_address)
            return True
        else:
            logger.error(
                "KEY CHANGED for %s! Stored: %s, Received: %s",
                device_address,
                stored_hex[:16] + "...",
                pubkey_hex[:16] + "...",
            )
            return False
    else:
        # First use — store the key
        known[device_address] = pubkey_hex
        _save_known_keys(known_keys_path, known)
        logger.info("Stored new key for %s (TOFU)", device_address)
        return True


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
    """Save known keys to JSON file."""
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "w") as f:
        json.dump(known, f, indent=2)
