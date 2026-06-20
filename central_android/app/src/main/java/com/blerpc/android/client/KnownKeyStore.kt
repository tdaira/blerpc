package com.blerpc.android.client

import android.content.Context

/**
 * TOFU (Trust On First Use) store for peripheral Ed25519 identity keys.
 *
 * The E2E handshake signature binds only the ephemeral X25519 keys, not the
 * peripheral's long-term identity key, so MitM resistance depends on the
 * central pinning that identity: on first connection the key is stored; on
 * later connections a changed key is rejected.
 *
 * Backed by a private SharedPreferences file (per-app, survives restarts).
 */
class KnownKeyStore(context: Context) {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    /**
     * First use → store the key and trust it. Subsequent connections → trust
     * only if the key matches the stored one; returns false on a TOFU violation
     * (the peripheral presented a different identity than the pinned one).
     */
    fun checkOrStore(
        deviceAddress: String,
        ed25519Pubkey: ByteArray,
    ): Boolean {
        val hex = ed25519Pubkey.joinToString("") { "%02x".format(it) }
        return when (val stored = prefs.getString(deviceAddress, null)) {
            null -> {
                prefs.edit().putString(deviceAddress, hex).apply()
                true
            }
            hex -> true
            else -> false
        }
    }

    companion object {
        private const val PREFS_NAME = "blerpc_known_keys"
    }
}
