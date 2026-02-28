import { BleManager, Device, Characteristic, Subscription } from 'react-native-ble-plx';
import { Buffer } from 'buffer';

export const SERVICE_UUID = '12340001-0000-1000-8000-00805f9b34fb';
export const CHAR_UUID = '12340002-0000-1000-8000-00805f9b34fb';

export interface ScannedDevice {
  device: Device;
  name: string | null;
  address: string;
}

export class BleTransport {
  private _manager: BleManager;
  private _device: Device | null = null;
  private _char: Characteristic | null = null;
  private _notifySub: Subscription | null = null;
  private _notifyQueue: Uint8Array[] = [];
  private _notifyWaiter: {
    resolve: (value: Uint8Array) => void;
    reject: (reason: Error) => void;
  } | null = null;
  private _mtu = 23;

  constructor() {
    this._manager = new BleManager();
  }

  get mtu(): number {
    return this._mtu;
  }

  async scan(timeout = 5000): Promise<ScannedDevice[]> {
    const allDevices = new Map<string, Device>();

    return new Promise<ScannedDevice[]>((resolve) => {
      this._manager.startDeviceScan(null, null, (error, device) => {
        if (error) {
          console.warn('[BLE] Scan error:', error.message);
          return;
        }
        if (device) {
          allDevices.set(device.id, device);
        }
      });

      setTimeout(() => {
        this._manager.stopDeviceScan();
        console.log(`[BLE] Scan complete: ${allDevices.size} device(s) found`);

        const devices: ScannedDevice[] = [];
        for (const device of allDevices.values()) {
          const serviceUUIDs = device.serviceUUIDs ?? [];
          const hasService = serviceUUIDs.some(
            (u) => u.toLowerCase() === SERVICE_UUID.toLowerCase(),
          );
          if (!hasService) continue;
          devices.push({
            device,
            name: device.name ?? device.localName ?? null,
            address: device.id,
          });
        }
        resolve(devices);
      }, timeout);
    });
  }

  async connect(scannedDevice: ScannedDevice): Promise<void> {
    this._device = await scannedDevice.device.connect({ requestMTU: 247 });

    // Get negotiated MTU
    this._mtu = this._device.mtu ?? 23;

    // Discover services and characteristics
    await this._device.discoverAllServicesAndCharacteristics();
    const services = await this._device.services();
    let targetChar: Characteristic | null = null;

    for (const service of services) {
      if (service.uuid.toLowerCase() === SERVICE_UUID.toLowerCase()) {
        const chars = await service.characteristics();
        for (const c of chars) {
          if (c.uuid.toLowerCase() === CHAR_UUID.toLowerCase()) {
            targetChar = c;
            break;
          }
        }
      }
    }

    if (!targetChar) {
      throw new Error('blerpc characteristic not found');
    }
    this._char = targetChar;

    // Enable notifications with queue
    this._notifyQueue = [];
    this._notifyWaiter = null;
    this._notifySub = this._char.monitor((error, characteristic) => {
      if (error) {
        console.warn('[BLE] Notify error:', error.message);
        return;
      }
      if (!characteristic?.value) return;

      const data = new Uint8Array(Buffer.from(characteristic.value, 'base64'));
      if (this._notifyWaiter) {
        const waiter = this._notifyWaiter;
        this._notifyWaiter = null;
        waiter.resolve(data);
      } else {
        this._notifyQueue.push(data);
      }
    });
  }

  async write(data: Uint8Array): Promise<void> {
    if (!this._char) throw new Error('Not connected');
    const base64 = Buffer.from(data).toString('base64');
    await this._char.writeWithoutResponse(base64);
  }

  async readNotify(timeout = 2000): Promise<Uint8Array> {
    if (!this._char) throw new Error('Not connected');
    if (this._notifyQueue.length > 0) {
      return this._notifyQueue.shift()!;
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
      // Done draining
    }
  }

  disconnect(): void {
    this._notifySub?.remove();
    this._notifySub = null;
    this._notifyQueue = [];
    this._notifyWaiter = null;
    if (this._device) {
      this._device.cancelConnection().catch(() => {});
      this._device = null;
    }
    this._char = null;
  }
}
