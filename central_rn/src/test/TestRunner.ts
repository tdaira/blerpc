import { ScannedDevice } from '../ble/BleTransport';
import { BlerpcClient } from '../client/BlerpcClient';

export type LogCallback = (message: string) => void;

export class TestRunner {
  private readonly _log: LogCallback;
  private _running = false;
  private _passCount = 0;
  private _failCount = 0;

  constructor(log: LogCallback) {
    this._log = log;
  }

  get isRunning(): boolean {
    return this._running;
  }

  async runAll(options: { iterations?: number; device?: ScannedDevice } = {}): Promise<void> {
    if (this._running) return;
    this._running = true;
    this._passCount = 0;
    this._failCount = 0;

    const iterations = options.iterations ?? 1;
    const client = new BlerpcClient();

    try {
      let target: ScannedDevice;
      if (options.device) {
        target = options.device;
      } else {
        this._log('Scanning for bleRPC peripherals...');
        const devices = await client.scan();
        if (devices.length === 0) {
          this._log('[ERROR] No bleRPC devices found');
          this._running = false;
          return;
        }
        target = devices[0];
      }
      this._log(`Connecting to ${target.name ?? target.address}...`);
      await client.connect(target);
      this._log(`Connected. MTU=${client.mtu}, encrypted=${client.isEncrypted}`);

      for (let iter = 1; iter <= iterations; iter++) {
        if (iterations > 1) this._log(`--- Iteration ${iter}/${iterations} ---`);

        await this._runTest(client, 'echo_basic', async () => {
          const resp = await client.echo({ message: 'hello' });
          this._check(resp.message === 'hello', `Expected 'hello', got '${resp.message}'`);
        });

        await this._runTest(client, 'echo_empty', async () => {
          const resp = await client.echo({ message: '' });
          this._check(resp.message === '', `Expected empty, got '${resp.message}'`);
        });

        await this._runTest(client, 'flash_read_basic', async () => {
          const resp = await client.flashRead({ address: 0, length: 64 });
          this._check(resp.data.length === 64, `Expected 64 bytes, got ${resp.data.length}`);
        });

        await this._runTest(client, 'flash_read_8kb', async () => {
          const resp = await client.flashRead({ address: 0, length: 8192 });
          this._check(resp.data.length === 8192, `Expected 8192 bytes, got ${resp.data.length}`);
        });

        await this._runTest(client, 'data_write', async () => {
          const testData = new Uint8Array(64);
          for (let i = 0; i < 64; i++) testData[i] = i;
          const resp = await client.dataWrite({ data: testData });
          this._check(resp.length === 64, `Expected length 64, got ${resp.length}`);
        });

        await this._runTest(client, 'counter_stream', async () => {
          const results = await client.counterStreamAll(5);
          this._check(results.length === 5, `Expected 5 results, got ${results.length}`);
          for (let i = 0; i < 5; i++) {
            this._check(results[i][0] === i, `Expected seq=${i}, got ${results[i][0]}`);
            this._check(results[i][1] === i * 10, `Expected value=${i * 10}, got ${results[i][1]}`);
          }
        });

        await this._runTest(client, 'counter_upload', async () => {
          const resp = await client.counterUploadAll(5);
          this._check(
            resp.receivedCount === 5,
            `Expected received_count=5, got ${resp.receivedCount}`,
          );
        });
      }

      this._log(
        `=== Functional: ${this._passCount} passed, ${this._failCount} failed (${iterations} iterations) ===`,
      );

      // Throughput benchmarks
      this._log('');
      this._log('=== Throughput Benchmarks ===');
      await this._benchFlashReadThroughput(client);
      await this._benchFlashReadOverhead(client);
      await this._benchEchoRoundtrip(client);
      await this._benchDataWriteThroughput(client);
      await this._benchStreamThroughput(client);
    } catch (e) {
      this._log(`[ERROR] ${e}`);
    } finally {
      client.disconnect();
      this._running = false;
    }
  }

  private async _benchFlashReadThroughput(client: BlerpcClient): Promise<void> {
    const readSize = 8192;
    const count = 10;
    const totalBytes = readSize * count;

    // Warmup
    await client.flashRead({ address: 0, length: readSize });

    const start = Date.now();
    for (let i = 0; i < count; i++) {
      const resp = await client.flashRead({ address: 0, length: readSize });
      this._check(resp.data.length === readSize, 'flash_read size mismatch');
    }
    const elapsedMs = Date.now() - start;
    const kbPerSec = totalBytes / 1024.0 / (elapsedMs / 1000.0);
    const msPerCall = elapsedMs / count;
    this._log(
      `[BENCH] flash_read_throughput: ${kbPerSec.toFixed(1)} KB/s (${totalBytes} bytes in ${elapsedMs}ms, ${msPerCall.toFixed(1)} ms/call)`,
    );
  }

  private async _benchFlashReadOverhead(client: BlerpcClient): Promise<void> {
    const count = 20;

    await client.flashRead({ address: 0, length: 1 });

    const start = Date.now();
    for (let i = 0; i < count; i++) {
      await client.flashRead({ address: 0, length: 1 });
    }
    const elapsedMs = Date.now() - start;
    const msPerCall = elapsedMs / count;
    this._log(
      `[BENCH] flash_read_overhead: ${msPerCall.toFixed(1)} ms/call (1 byte x ${count} calls in ${elapsedMs}ms)`,
    );
  }

  private async _benchEchoRoundtrip(client: BlerpcClient): Promise<void> {
    const count = 50;

    await client.echo({ message: 'x' });

    const start = Date.now();
    for (let i = 0; i < count; i++) {
      await client.echo({ message: 'hello' });
    }
    const elapsedMs = Date.now() - start;
    const msPerCall = elapsedMs / count;
    this._log(
      `[BENCH] echo_roundtrip: ${msPerCall.toFixed(1)} ms/call (${count} calls in ${elapsedMs}ms)`,
    );
  }

  private async _benchDataWriteThroughput(client: BlerpcClient): Promise<void> {
    const writeSize = 200;
    const count = 20;
    const totalBytes = writeSize * count;
    const testData = new Uint8Array(writeSize);
    for (let i = 0; i < writeSize; i++) testData[i] = i % 256;

    await client.dataWrite({ data: testData });

    const start = Date.now();
    for (let i = 0; i < count; i++) {
      await client.dataWrite({ data: testData });
    }
    const elapsedMs = Date.now() - start;
    const kbPerSec = totalBytes / 1024.0 / (elapsedMs / 1000.0);
    const msPerCall = elapsedMs / count;
    this._log(
      `[BENCH] data_write_throughput: ${kbPerSec.toFixed(1)} KB/s (${totalBytes} bytes in ${elapsedMs}ms, ${msPerCall.toFixed(1)} ms/call)`,
    );
  }

  private async _benchStreamThroughput(client: BlerpcClient): Promise<void> {
    const count = 20;

    const start1 = Date.now();
    const results = await client.counterStreamAll(count);
    const elapsed1 = Date.now() - start1;
    this._check(results.length === count, 'stream count mismatch');
    this._log(
      `[BENCH] counter_stream (P->C): ${count} items in ${elapsed1}ms (${(elapsed1 / count).toFixed(1)} ms/item)`,
    );

    const start2 = Date.now();
    const resp = await client.counterUploadAll(count);
    const elapsed2 = Date.now() - start2;
    this._check(resp.receivedCount === count, 'upload count mismatch');
    this._log(
      `[BENCH] counter_upload (C->P): ${count} items in ${elapsed2}ms (${(elapsed2 / count).toFixed(1)} ms/item)`,
    );
  }

  private async _runTest(
    client: BlerpcClient,
    name: string,
    block: () => Promise<void>,
  ): Promise<void> {
    try {
      await block();
      this._passCount++;
      this._log(`[PASS] ${name}`);
    } catch (e) {
      this._failCount++;
      this._log(`[FAIL] ${name}: ${e}`);
      await new Promise((resolve) => setTimeout(resolve, 500));
      await client.transport.drainNotifications();
    }
  }

  private _check(condition: boolean, message: string): void {
    if (!condition) throw new Error(message);
  }
}
