import AsyncStorage from '@react-native-async-storage/async-storage';

// TOFU (Trust On First Use) store for peripheral Ed25519 identity keys.
//
// The E2E handshake signature binds only the ephemeral X25519 keys, not the
// peripheral's long-term identity key, so MitM resistance depends on the
// central pinning that identity: on first connection the key is stored; on
// later connections a changed key is rejected.
//
// Backed by AsyncStorage (per-app, survives restarts). Because the verify
// callback is synchronous, the map is loaded before the key exchange and
// persisted after it.

const STORE_KEY = 'blerpc_known_keys';

export type KnownKeys = Record<string, string>;

export async function loadKnownKeys(): Promise<KnownKeys> {
  try {
    const raw = await AsyncStorage.getItem(STORE_KEY);
    return raw ? (JSON.parse(raw) as KnownKeys) : {};
  } catch {
    return {};
  }
}

export async function saveKnownKeys(known: KnownKeys): Promise<void> {
  try {
    await AsyncStorage.setItem(STORE_KEY, JSON.stringify(known));
  } catch {
    // Best-effort persistence; pinning still holds for the current session.
  }
}

function toHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * First use → store the key in `known` and trust it. Subsequent connections →
 * trust only if the key matches; returns false on a TOFU violation (the
 * peripheral presented a different identity than the pinned one).
 */
export function checkOrStore(
  known: KnownKeys,
  address: string,
  ed25519Pubkey: Uint8Array,
): boolean {
  const hex = toHex(ed25519Pubkey);
  const stored = known[address];
  if (stored === undefined) {
    known[address] = hex;
    return true;
  }
  return stored === hex;
}
