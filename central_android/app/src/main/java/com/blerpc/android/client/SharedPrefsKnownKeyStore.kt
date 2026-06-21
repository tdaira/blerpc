package com.blerpc.android.client

import android.content.Context
import com.blerpc.protocol.KnownKeyStore

/**
 * SharedPreferences-backed [KnownKeyStore] for TOFU identity pinning.
 *
 * The pinning policy (trust on first use, reject a changed key) lives in the
 * protocol library (`tofuVerify`); this class only provides per-app persistence
 * that survives restarts.
 */
class SharedPrefsKnownKeyStore(context: Context) : KnownKeyStore {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    override fun get(deviceId: String): String? = prefs.getString(deviceId, null)

    override fun put(
        deviceId: String,
        hexEd25519Pubkey: String,
    ) {
        prefs.edit().putString(deviceId, hexEd25519Pubkey).apply()
    }

    companion object {
        private const val PREFS_NAME = "blerpc_known_keys"
    }
}
