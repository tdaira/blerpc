import 'dart:typed_data';

import 'package:shared_preferences/shared_preferences.dart';

/// TOFU (Trust On First Use) store for peripheral Ed25519 identity keys.
///
/// The E2E handshake signature binds only the ephemeral X25519 keys, not the
/// peripheral's long-term identity key, so MitM resistance depends on the
/// central pinning that identity: on first connection the key is stored; on
/// later connections a changed key is rejected.
///
/// Backed by [SharedPreferences] (per-app, survives restarts).
class KnownKeyStore {
  KnownKeyStore(this._prefs);

  final SharedPreferences _prefs;
  static const String _prefix = 'blerpc_known_key_';

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
    final stored = _prefs.getString('$_prefix$address');
    if (stored == null) {
      _prefs.setString('$_prefix$address', hex); // persisted asynchronously
      return true;
    }
    return stored == hex;
  }
}
