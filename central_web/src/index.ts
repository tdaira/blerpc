// blerpc central for the browser (Web Bluetooth). Bundled to
// `dist/blerpc-web.js` (global `BlerpcWeb`) and `dist/blerpc-web.mjs` (ESM).

export {
  BlerpcClient,
  PayloadTooLargeError,
  ResponseTooLargeError,
  PeripheralErrorException,
  ProtocolException,
} from './client/BlerpcClient';
export { WebBluetoothTransport, SERVICE_UUID, CHAR_UUID } from './ble/WebBluetoothTransport';
export type { ScannedDevice, WebBluetoothTransportOptions } from './ble/WebBluetoothTransport';
export { LocalStorageKnownKeyStore } from './client/knownKeys';
export { blerpc } from './proto/blerpc';
