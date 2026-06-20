import Foundation

/// TOFU (Trust On First Use) store for peripheral Ed25519 identity keys.
///
/// The E2E handshake signature binds only the ephemeral X25519 keys, not the
/// peripheral's long-term identity key, so MitM resistance depends on the
/// central pinning that identity: on first connection the key is stored; on
/// later connections a changed key is rejected.
///
/// Backed by `UserDefaults` (per-app, survives restarts).
final class KnownKeyStore {
    private let defaults: UserDefaults
    private let storeKey = "blerpc_known_keys"

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    /// First use → store the key and trust it. Subsequent connections → trust
    /// only if the key matches the stored one; returns false on a TOFU violation
    /// (the peripheral presented a different identity than the pinned one).
    func checkOrStore(deviceId: String, ed25519Pubkey: Data) -> Bool {
        let hex = ed25519Pubkey.map { String(format: "%02x", $0) }.joined()
        var known = defaults.dictionary(forKey: storeKey) as? [String: String] ?? [:]
        if let stored = known[deviceId] {
            return stored == hex
        }
        known[deviceId] = hex
        defaults.set(known, forKey: storeKey)
        return true
    }
}
