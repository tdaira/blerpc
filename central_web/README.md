# central_web — browser (Web Bluetooth) blerpc central

A bleRPC central that runs in the browser over the [Web Bluetooth API]. It reuses the
platform-neutral protocol library [`@blerpc/protocol-ts`] (container framing, command packets,
and the X25519/Ed25519/AES-128-GCM encryption layer run unmodified in a browser) and adds a
single new piece: a `WebBluetoothTransport`. The protocol logic in `BlerpcClient` is the same as
the React Native client (`central_rn`); the transport is injected so one client drives both.

## Layout

```
central_web/
├── src/
│   ├── ble/WebBluetoothTransport.ts  # navigator.bluetooth GATT transport (the only new code)
│   ├── client/BlerpcClient.ts        # protocol client (transport injected)
│   ├── client/GeneratedClient.ts     # generated RPC methods (from proto/blerpc.proto)
│   ├── client/knownKeys.ts           # localStorage-backed TOFU KnownKeyStore
│   ├── proto/blerpc.{js,d.ts}        # protobufjs static module (pbjs/pbts)
│   └── index.ts                      # bundle entry / public API
├── index.html                        # minimal connect + echo demo
├── build.mjs                         # esbuild → dist/blerpc-web.{js,mjs}
├── package.json
└── tsconfig.json
```

## Build & run

```bash
npm install
npm run proto      # regenerate src/proto/blerpc.* from ../proto/blerpc.proto (optional)
npm run typecheck
npm run build      # -> dist/blerpc-web.js (global BlerpcWeb) + dist/blerpc-web.mjs (ESM)
npx serve .        # serve over http://localhost (Web Bluetooth needs a secure context)
# open http://localhost:3000/ in Chrome/Edge, click Connect
```

## Usage

Plain `<script>` (global `BlerpcWeb`):

```html
<script src="./dist/blerpc-web.js"></script>
<script>
  const client = new BlerpcWeb.BlerpcClient();
  const devices = await client.scan();     // opens the browser device chooser (needs a user gesture)
  await client.connect(devices[0]);        // negotiation + key exchange
  const resp = await client.echo({ message: 'hello' });
</script>
```

ESM:

```js
import { BlerpcClient } from './dist/blerpc-web.mjs';
```

## Notes / constraints

- **Chromium only** (Chrome/Edge/Opera), **secure context** (HTTPS or `localhost`), and
  `scan()`/`requestDevice` must be called from a **user gesture**.
- Web Bluetooth exposes **no MTU getter** and **no passive scan**: the MTU is a
  `WebBluetoothTransport` constructor option (default 247), and `scan()` returns the single device
  the user picks in the chooser.
- Encryption + TOFU identity pinning are on by default (`new BlerpcClient(transport, true, true)`).
  Pinned peripheral identities are stored in `localStorage` under `blerpc_known_keys`.

[Web Bluetooth API]: https://developer.mozilla.org/en-US/docs/Web/API/Web_Bluetooth_API
[`@blerpc/protocol-ts`]: https://github.com/tdaira/blerpc-protocol-ts
