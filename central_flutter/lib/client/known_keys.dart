import 'dart:convert';

import 'package:blerpc_protocol/blerpc_protocol.dart' show KnownKeyStore;
import 'package:shared_preferences/shared_preferences.dart';

/// SharedPreferences-backed [KnownKeyStore] for TOFU identity pinning.
///
/// The pinning policy (trust on first use, reject a changed key) lives in the
/// protocol library (`tofuVerify`); this class only provides per-app
/// persistence that survives restarts. The map is loaded before the key
/// exchange so `get`/`put` run synchronously inside the library, and [persist]
/// is awaited afterwards to flush newly pinned keys to disk.
class SharedPrefsKnownKeyStore implements KnownKeyStore {
  SharedPrefsKnownKeyStore(this._prefs) {
    final raw = _prefs.getString(_storeKey);
    _keys = raw == null
        ? <String, String>{}
        : (jsonDecode(raw) as Map).cast<String, String>();
  }

  final SharedPreferences _prefs;
  static const String _storeKey = 'blerpc_known_keys';
  late final Map<String, String> _keys;
  bool _dirty = false;

  /// Load the store. Call before the key exchange.
  static Future<SharedPrefsKnownKeyStore> load() async {
    return SharedPrefsKnownKeyStore(await SharedPreferences.getInstance());
  }

  @override
  String? get(String deviceId) => _keys[deviceId];

  @override
  void put(String deviceId, String hexEd25519Pubkey) {
    _keys[deviceId] = hexEd25519Pubkey;
    _dirty = true;
  }

  /// Flush newly pinned keys to disk. Call after a successful key exchange.
  Future<void> persist() async {
    if (_dirty) {
      await _prefs.setString(_storeKey, jsonEncode(_keys));
      _dirty = false;
    }
  }
}
