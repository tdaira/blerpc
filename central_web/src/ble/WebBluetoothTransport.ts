// Web Bluetooth (browser) transport for blerpc.
//
// Implements the same transport contract the RN `BleTransport` exposes
// (`mtu`, `scan`, `connect`, `write`, `readNotify`, `drainNotifications`,
// `disconnect`) so `BlerpcClient` can drive it unchanged. The only real
// differences from the RN transport are platform ones:
//
//   - `navigator.bluetooth.requestDevice` shows the browser's device chooser
//     and returns a single device — there is no passive multi-device scan, so
//     `scan()` returns a one-element list (the chosen device). It MUST be
//     called from a user gesture (click/tap) and only in a secure context
//     (HTTPS or localhost). Chromium-only (Chrome/Edge/Opera).
//   - Web Bluetooth exposes no MTU getter. Chromium negotiates up to 247
//     automatically; the MTU is a constructor option (default 247) and the
//     `ContainerSplitter` keeps every write within `effectiveMtu = mtu - 3`.
//   - Notifications arrive as `DataView`s (no base64/Buffer dance).

export const SERVICE_UUID = '12340001-0000-1000-8000-00805f9b34fb';
export const CHAR_UUID = '12340002-0000-1000-8000-00805f9b34fb';

export interface ScannedDevice {
  device: BluetoothDevice;
  name: string | null;
  address: string;
  rssi: number;
}

export interface WebBluetoothTransportOptions {
  /** ATT MTU to assume (no getter in Web Bluetooth). Default 247. */
  mtu?: number;
  /** Extra name prefix to offer in the chooser (default "Orphe"). */
  namePrefix?: string;
}

export class WebBluetoothTransport {
  private _device: BluetoothDevice | null = null;
  private _char: BluetoothRemoteGATTCharacteristic | null = null;
  private _onValue: ((ev: Event) => void) | null = null;
  private _notifyQueue: Uint8Array[] = [];
  private _notifyWaiter: {
    resolve: (value: Uint8Array) => void;
    reject: (reason: Error) => void;
  } | null = null;
  private readonly _mtu: number;
  private readonly _namePrefix: string;

  constructor(options: WebBluetoothTransportOptions = {}) {
    this._mtu = options.mtu ?? 247;
    this._namePrefix = options.namePrefix ?? 'Orphe';
  }

  get mtu(): number {
    return this._mtu;
  }

  /**
   * Open the browser device chooser and return the selected device as a
   * one-element list. `timeout` is accepted for API parity but ignored — the
   * chooser is driven by the user, not a scan window. Must run inside a user
   * gesture and a secure context.
   */
  async scan(_timeout = 5000): Promise<ScannedDevice[]> {
    if (typeof navigator === 'undefined' || !navigator.bluetooth) {
      throw new Error('Web Bluetooth is not available (use Chrome/Edge over HTTPS or localhost)');
    }
    const device = await navigator.bluetooth.requestDevice({
      filters: [{ services: [SERVICE_UUID] }, { namePrefix: this._namePrefix }],
      optionalServices: [SERVICE_UUID],
    });
    return [
      {
        device,
        name: device.name ?? null,
        address: device.id,
        rssi: 0, // not exposed by Web Bluetooth
      },
    ];
  }

  async connect(scannedDevice: ScannedDevice): Promise<void> {
    const device = scannedDevice.device;
    this._device = device;
    if (!device.gatt) throw new Error('Device has no GATT server');

    const server = await device.gatt.connect();
    const service = await server.getPrimaryService(SERVICE_UUID);
    this._char = await service.getCharacteristic(CHAR_UUID);

    this._notifyQueue = [];
    this._notifyWaiter = null;
    this._onValue = (ev: Event) => {
      const target = ev.target as BluetoothRemoteGATTCharacteristic;
      const dv = target.value;
      if (!dv) return;
      // Copy out of the shared buffer — Web Bluetooth may reuse it.
      const data = new Uint8Array(dv.byteLength);
      for (let i = 0; i < dv.byteLength; i++) data[i] = dv.getUint8(i);
      if (this._notifyWaiter) {
        const waiter = this._notifyWaiter;
        this._notifyWaiter = null;
        waiter.resolve(data);
      } else {
        this._notifyQueue.push(data);
      }
    };
    this._char.addEventListener('characteristicvaluechanged', this._onValue);
    await this._char.startNotifications();
  }

  async write(data: Uint8Array): Promise<void> {
    if (!this._char) throw new Error('Not connected');
    // Write-without-response — same as the RN transport's writeWithoutResponse.
    // A copy keeps the underlying ArrayBuffer sized exactly to the payload.
    const buf = new Uint8Array(data);
    if (this._char.writeValueWithoutResponse) {
      await this._char.writeValueWithoutResponse(buf);
    } else {
      // Older Chromium exposes only the deprecated writeValue().
      await this._char.writeValue(buf);
    }
  }

  async readNotify(timeout = 2000): Promise<Uint8Array> {
    if (!this._char) throw new Error('Not connected');
    if (this._notifyQueue.length > 0) {
      return this._notifyQueue.shift() as Uint8Array;
    }
    return new Promise<Uint8Array>((resolve, reject) => {
      const timer = setTimeout(() => {
        this._notifyWaiter = null;
        reject(new Error('Timeout waiting for notification'));
      }, timeout);
      this._notifyWaiter = {
        resolve: (value) => {
          clearTimeout(timer);
          resolve(value);
        },
        reject: (reason) => {
          clearTimeout(timer);
          reject(reason);
        },
      };
    });
  }

  async drainNotifications(): Promise<void> {
    this._notifyQueue = [];
    try {
      for (;;) {
        await this.readNotify(100);
      }
    } catch {
      // Done draining.
    }
  }

  disconnect(): void {
    if (this._char && this._onValue) {
      this._char.removeEventListener('characteristicvaluechanged', this._onValue);
    }
    this._onValue = null;
    this._notifyQueue = [];
    this._notifyWaiter = null;
    if (this._device?.gatt?.connected) {
      this._device.gatt.disconnect();
    }
    this._device = null;
    this._char = null;
  }
}
