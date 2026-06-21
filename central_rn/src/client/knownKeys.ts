import AsyncStorage from '@react-native-async-storage/async-storage';
import type { KnownKeyStore } from '@blerpc/protocol-rn';

// TOFU (Trust On First Use) store for peripheral Ed25519 identity keys.
//
// The pinning policy (trust on first use, reject a changed key) lives in the
// protocol library (`tofuVerify`); this class only provides per-app persistence.
// Backed by AsyncStorage (survives restarts). Because the library reads the
// store synchronously during the key exchange, the map is loaded before the
// exchange and persisted after it.

const STORE_KEY = 'blerpc_known_keys';

export class AsyncStorageKnownKeyStore implements KnownKeyStore {
  private keys: Record<string, string>;
  private dirty = false;

  private constructor(keys: Record<string, string>) {
    this.keys = keys;
  }

  /** Load the store. Call before the key exchange. */
  static async load(): Promise<AsyncStorageKnownKeyStore> {
    let keys: Record<string, string> = {};
    try {
      const raw = await AsyncStorage.getItem(STORE_KEY);
      if (raw) keys = JSON.parse(raw) as Record<string, string>;
    } catch {
      // Best-effort load; an empty store just means everything is first-use.
    }
    return new AsyncStorageKnownKeyStore(keys);
  }

  get(deviceId: string): string | null {
    return this.keys[deviceId] ?? null;
  }

  put(deviceId: string, hexEd25519Pubkey: string): void {
    this.keys[deviceId] = hexEd25519Pubkey;
    this.dirty = true;
  }

  /** Flush newly pinned keys to disk. Call after a successful key exchange. */
  async persist(): Promise<void> {
    if (!this.dirty) return;
    try {
      await AsyncStorage.setItem(STORE_KEY, JSON.stringify(this.keys));
      this.dirty = false;
    } catch {
      // Best-effort persistence; pinning still holds for the current session.
    }
  }
}
