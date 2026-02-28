import {
  ContainerSplitter,
  ContainerAssembler,
  Container,
  ContainerType,
  ControlCmd,
  CommandPacket,
  CommandType,
  makeTimeoutRequest,
  makeCapabilitiesRequest,
  makeStreamEndC2P,
  makeKeyExchange,
  centralPerformKeyExchange,
  BlerpcCryptoSession,
  CAPABILITY_FLAG_ENCRYPTION_SUPPORTED,
  BLERPC_ERROR_RESPONSE_TOO_LARGE,
} from '@blerpc/protocol-rn';
import { BleTransport, ScannedDevice } from '../ble/BleTransport';
import { GeneratedClient } from './GeneratedClient';

export class PayloadTooLargeError extends Error {
  constructor(
    public actual: number,
    public limit: number,
  ) {
    super(`Request payload (${actual} bytes) exceeds peripheral limit (${limit} bytes)`);
  }
}

export class ResponseTooLargeError extends Error {
  constructor(message: string) {
    super(`ResponseTooLargeError: ${message}`);
  }
}

export class PeripheralErrorException extends Error {
  constructor(public errorCode: number) {
    super(`PeripheralErrorException: 0x${errorCode.toString(16).padStart(2, '0')}`);
  }
}

export class ProtocolException extends Error {
  constructor(message: string) {
    super(`ProtocolException: ${message}`);
  }
}

export class BlerpcClient extends GeneratedClient {
  readonly transport = new BleTransport();
  private readonly requireEncryption: boolean;

  private _splitter: ContainerSplitter | null = null;
  private readonly _assembler = new ContainerAssembler();
  private _timeout = 100;
  private _maxRequestPayloadSize: number | null = null;

  // Encryption state
  private _session: BlerpcCryptoSession | null = null;

  constructor(requireEncryption = true) {
    super();
    this.requireEncryption = requireEncryption;
  }

  get mtu(): number {
    return this.transport.mtu;
  }

  get isEncrypted(): boolean {
    return this._session !== null;
  }

  private _readTimeout(firstRead: boolean): number {
    if (!firstRead) return this._timeout;
    return this._timeout > 2000 ? this._timeout : 2000;
  }

  private _handleControlError(container: Container): void {
    if (container.controlCmd === ControlCmd.ERROR && container.payload.length > 0) {
      const errorCode = container.payload[0];
      if (errorCode === BLERPC_ERROR_RESPONSE_TOO_LARGE) {
        throw new ResponseTooLargeError("Response exceeds peripheral's max_response_payload_size");
      }
      throw new PeripheralErrorException(errorCode);
    }
  }

  async scan(timeout?: number): Promise<ScannedDevice[]> {
    return this.transport.scan(timeout ?? 5000);
  }

  async connect(device: ScannedDevice): Promise<void> {
    await this.transport.connect(device);
    this._splitter = new ContainerSplitter(this.transport.mtu);

    try {
      await this._requestTimeout();
    } catch {
      console.log('Peripheral did not respond to timeout request, using default');
    }
    try {
      await this._requestCapabilities();
    } catch {
      console.log('Peripheral did not respond to capabilities request');
    }

    if (this.requireEncryption && this._session === null) {
      throw new Error(
        'Encryption required but key exchange was not completed. ' +
          'The peripheral may not support encryption or a MitM may ' +
          'have stripped the encryption capability flag.',
      );
    }
  }

  private async _requestTimeout(): Promise<void> {
    const s = this._splitter!;
    const tid = s.nextTransactionId();
    const req = makeTimeoutRequest(tid);
    await this.transport.write(req.serialize());
    const data = await this.transport.readNotify(1000);
    const resp = Container.deserialize(data);
    if (
      resp.containerType === ContainerType.CONTROL &&
      resp.controlCmd === ControlCmd.TIMEOUT &&
      resp.payload.length === 2
    ) {
      const bd = new DataView(
        resp.payload.buffer,
        resp.payload.byteOffset,
        resp.payload.byteLength,
      );
      const ms = bd.getUint16(0, true);
      this._timeout = ms;
      console.log(`Peripheral timeout: ${ms}ms`);
    }
  }

  private async _requestCapabilities(): Promise<void> {
    const s = this._splitter!;
    const tid = s.nextTransactionId();
    const req = makeCapabilitiesRequest(tid);
    await this.transport.write(req.serialize());
    const data = await this.transport.readNotify(1000);
    const resp = Container.deserialize(data);
    if (
      resp.containerType === ContainerType.CONTROL &&
      resp.controlCmd === ControlCmd.CAPABILITIES &&
      resp.payload.length >= 6
    ) {
      const bd = new DataView(
        resp.payload.buffer,
        resp.payload.byteOffset,
        resp.payload.byteLength,
      );
      const maxReq = bd.getUint16(0, true);
      const flags = bd.getUint16(4, true);
      this._maxRequestPayloadSize = maxReq;
      console.log(
        `Capabilities: max_req=${maxReq}, flags=0x${flags.toString(16).padStart(4, '0')}`,
      );

      if (flags & CAPABILITY_FLAG_ENCRYPTION_SUPPORTED) {
        await this._performKeyExchange();
      }
    }
  }

  private async _performKeyExchange(): Promise<void> {
    const s = this._splitter!;

    try {
      this._session = await centralPerformKeyExchange({
        send: async (payload: Uint8Array) => {
          const tid = s.nextTransactionId();
          const req = makeKeyExchange(tid, payload);
          await this.transport.write(req.serialize());
        },
        receive: async () => {
          const data = await this.transport.readNotify(2000);
          const resp = Container.deserialize(data);
          if (
            resp.containerType !== ContainerType.CONTROL ||
            resp.controlCmd !== ControlCmd.KEY_EXCHANGE
          ) {
            throw new Error('Expected KEY_EXCHANGE response, got something else');
          }
          return resp.payload;
        },
      });
      console.log('E2E encryption established');
    } catch (e) {
      console.log('Key exchange failed:', e);
      if (this.requireEncryption) throw e;
    }
  }

  private _encryptPayload(payload: Uint8Array): Uint8Array {
    if (this._session === null) {
      if (this.requireEncryption) {
        throw new Error('Encryption required but no session established');
      }
      return payload;
    }
    return this._session.encrypt(payload);
  }

  private _decryptPayload(payload: Uint8Array): Uint8Array {
    if (this._session === null) {
      if (this.requireEncryption) {
        throw new Error('Encryption required but no session established');
      }
      return payload;
    }
    return this._session.decrypt(payload);
  }

  protected async call(cmdName: string, requestData: Uint8Array): Promise<Uint8Array> {
    const s =
      this._splitter ??
      (() => {
        throw new Error('Not connected');
      })();

    const cmd = new CommandPacket({
      cmdType: CommandType.REQUEST,
      cmdName,
      data: requestData,
    });
    const payload = cmd.serialize();

    if (this._maxRequestPayloadSize !== null && payload.length > this._maxRequestPayloadSize) {
      throw new PayloadTooLargeError(payload.length, this._maxRequestPayloadSize);
    }

    const sendPayload = this._encryptPayload(payload);
    const containers = s.split(sendPayload);
    for (const c of containers) {
      await this.transport.write(c.serialize());
    }

    this._assembler.reset();
    let firstRead = true;
    for (;;) {
      const notifyData = await this.transport.readNotify(this._readTimeout(firstRead));
      firstRead = false;
      const container = Container.deserialize(notifyData);

      if (container.containerType === ContainerType.CONTROL) {
        this._handleControlError(container);
        continue;
      }

      const result = this._assembler.feed(container);
      if (result !== null) {
        const decrypted = this._decryptPayload(result);
        const resp = CommandPacket.deserialize(decrypted);
        if (resp.cmdType !== CommandType.RESPONSE) {
          throw new ProtocolException(`Expected response, got type=${resp.cmdType}`);
        }
        if (resp.cmdName !== cmdName) {
          throw new ProtocolException(
            `Command name mismatch: expected '${cmdName}', got '${resp.cmdName}'`,
          );
        }
        return resp.data;
      }
    }
  }

  protected async streamReceive(cmdName: string, requestData: Uint8Array): Promise<Uint8Array[]> {
    const s =
      this._splitter ??
      (() => {
        throw new Error('Not connected');
      })();

    const cmd = new CommandPacket({
      cmdType: CommandType.REQUEST,
      cmdName,
      data: requestData,
    });
    const payload = cmd.serialize();

    if (this._maxRequestPayloadSize !== null && payload.length > this._maxRequestPayloadSize) {
      throw new PayloadTooLargeError(payload.length, this._maxRequestPayloadSize);
    }

    const sendPayload = this._encryptPayload(payload);
    const containers = s.split(sendPayload);
    for (const c of containers) {
      await this.transport.write(c.serialize());
    }

    const results: Uint8Array[] = [];
    this._assembler.reset();
    let firstRead = true;
    for (;;) {
      const notifyData = await this.transport.readNotify(this._readTimeout(firstRead));
      firstRead = false;
      const container = Container.deserialize(notifyData);

      if (container.containerType === ContainerType.CONTROL) {
        if (container.controlCmd === ControlCmd.STREAM_END_P2C) break;
        this._handleControlError(container);
        continue;
      }

      const result = this._assembler.feed(container);
      if (result !== null) {
        const decrypted = this._decryptPayload(result);
        const resp = CommandPacket.deserialize(decrypted);
        if (resp.cmdType !== CommandType.RESPONSE) {
          throw new ProtocolException(`Expected response, got type=${resp.cmdType}`);
        }
        results.push(resp.data);
      }
    }
    return results;
  }

  protected async streamSend(
    cmdName: string,
    messages: Uint8Array[],
    finalCmdName: string,
  ): Promise<Uint8Array> {
    const s =
      this._splitter ??
      (() => {
        throw new Error('Not connected');
      })();

    for (const msgData of messages) {
      const cmd = new CommandPacket({
        cmdType: CommandType.REQUEST,
        cmdName,
        data: msgData,
      });
      const payload = cmd.serialize();
      const sendPayload = this._encryptPayload(payload);
      const containers = s.split(sendPayload);
      for (const c of containers) {
        await this.transport.write(c.serialize());
      }
    }

    // Send STREAM_END_C2P
    const tid = s.nextTransactionId();
    const streamEnd = makeStreamEndC2P(tid);
    await this.transport.write(streamEnd.serialize());

    // Wait for final response
    this._assembler.reset();
    let firstRead = true;
    for (;;) {
      const notifyData = await this.transport.readNotify(this._readTimeout(firstRead));
      firstRead = false;
      const container = Container.deserialize(notifyData);

      if (container.containerType === ContainerType.CONTROL) {
        this._handleControlError(container);
        continue;
      }

      const result = this._assembler.feed(container);
      if (result !== null) {
        const decrypted = this._decryptPayload(result);
        const resp = CommandPacket.deserialize(decrypted);
        if (resp.cmdType !== CommandType.RESPONSE) {
          throw new ProtocolException(`Expected response, got type=${resp.cmdType}`);
        }
        if (resp.cmdName !== finalCmdName) {
          throw new ProtocolException(
            `Command name mismatch: expected '${finalCmdName}', got '${resp.cmdName}'`,
          );
        }
        return resp.data;
      }
    }
  }

  disconnect(): void {
    this.transport.disconnect();
    this._session = null;
    this._splitter = null;
  }
}
