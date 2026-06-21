import BlerpcProtocol
import Foundation

/// `UserDefaults`-backed ``KnownKeyStore`` for TOFU identity pinning.
///
/// The pinning policy (trust on first use, reject a changed key) lives in the
/// protocol library (``tofuVerify(store:deviceId:ed25519Pubkey:)``); this type
/// only provides per-app persistence that survives restarts. `UserDefaults` is
/// synchronous, so `get`/`put` are read/written directly by the library during
/// the key exchange.
final class UserDefaultsKnownKeyStore: KnownKeyStore {
    private let defaults: UserDefaults
    private let storeKey = "blerpc_known_keys"

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    func get(deviceId: String) -> String? {
        let known = defaults.dictionary(forKey: storeKey) as? [String: String]
        return known?[deviceId]
    }

    func put(deviceId: String, hexEd25519Pubkey: String) {
        var known = defaults.dictionary(forKey: storeKey) as? [String: String] ?? [:]
        known[deviceId] = hexEd25519Pubkey
        defaults.set(known, forKey: storeKey)
    }
}
