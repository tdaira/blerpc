import 'dart:convert';
import 'dart:typed_data';

import 'package:shared_preferences/shared_preferences.dart';

/// TOFU (Trust On First Use) store for peripheral Ed25519 identity keys.
///
/// The E2E handshake signature binds only the ephemeral X25519 keys, not the
/// peripheral's long-term identity key, so MitM resistance depends on the
/// central pinning that identity: on first connection the key is stored; on
/// later connections a changed key is rejected.
///
/// Backed by [SharedPreferences] (per-app, survives restarts). The map is
/// loaded before the key exchange so [checkOrStore] can run synchronously
/// inside the verify callback; [persist] is awaited afterwards to flush it.
class KnownKeyStore {
  KnownKeyStore(this._prefs) {
    final raw = _prefs.getString(_storeKey);
    _keys = raw == null
        ? <String, String>{}
        : (jsonDecode(raw) as Map).cast<String, String>();
  }

  final SharedPreferences _prefs;
  static const String _storeKey = 'blerpc_known_keys';
  late final Map<String, String> _keys;
  bool _dirty = false;

  /// Load the store. Call before the key exchange so [checkOrStore] can run
  /// synchronously inside the verify callback.
  static Future<KnownKeyStore> load() async {
    return KnownKeyStore(await SharedPreferences.getInstance());
  }

  /// First use → store the key and trust it. Subsequent connections → trust
  /// only if the key matches the stored one; returns false on a TOFU violation
  /// (the peripheral presented a different identity than the pinned one).
  bool checkOrStore(String address, Uint8List ed25519Pubkey) {
    final hex =
        ed25519Pubkey.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
    final stored = _keys[address];
    if (stored == null) {
      _keys[address] = hex;
      _dirty = true;
      return true;
    }
    return stored == hex;
  }

  /// Flush newly pinned keys to disk. Call after a successful key exchange.
  Future<void> persist() async {
    if (_dirty) {
      await _prefs.setString(_storeKey, jsonEncode(_keys));
      _dirty = false;
    }
  }
}
