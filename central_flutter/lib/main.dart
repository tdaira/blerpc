import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter/services.dart';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';

import 'ble/ble_transport.dart';
import 'design/ble_colors.dart';
import 'design/ble_components.dart';
import 'design/ble_dimens.dart';
import 'design/ble_theme.dart';
import 'test/test_runner.dart';

const _autoRun = bool.fromEnvironment('AUTO_RUN', defaultValue: false);

void main() {
  runApp(const BlerpcCentralApp());
}

class BlerpcCentralApp extends StatelessWidget {
  const BlerpcCentralApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'bleRPC Central',
      theme: BleTheme.dark(),
      home: const HomePage(),
    );
  }
}

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final List<String> _logs = [];
  final ScrollController _scrollController = ScrollController();
  List<ScannedDevice> _devices = [];
  bool _scanning = false;
  bool _running = false;
  bool _autoRunStarted = false;
  bool _showCopied = false;
  late final TestRunner _testRunner;

  @override
  void initState() {
    super.initState();
    _testRunner = TestRunner(log: _addLog);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_autoRun && !_autoRunStarted) {
      _autoRunStarted = true;
      Future.delayed(const Duration(milliseconds: 500), _autoScanAndRun);
    }
  }

  Future<void> _autoScanAndRun() async {
    debugPrint('[AUTO] Starting auto scan and run...');
    await _scan();
    if (_devices.isNotEmpty) {
      debugPrint('[AUTO] Found ${_devices.length} device(s), running tests...');
      await _runTests(_devices.first);
    } else {
      debugPrint('[AUTO] No devices found');
    }
  }

  void _addLog(String msg) {
    debugPrint(msg);
    setState(() {
      _logs.add(msg);
    });
    SchedulerBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 100),
          curve: Curves.easeOut,
        );
      }
    });
  }

  Future<void> _scan() async {
    setState(() {
      _scanning = true;
      _devices = [];
      _logs.clear();
    });
    _addLog('Scanning...');

    try {
      // On iOS the initial adapter state is often "unknown" until
      // CoreBluetooth finishes initialising.  Wait up to 5 seconds for
      // the state to settle to something other than "unknown".
      final state = await FlutterBluePlus.adapterState
          .where((s) => s != BluetoothAdapterState.unknown)
          .first
          .timeout(const Duration(seconds: 5),
              onTimeout: () => BluetoothAdapterState.unknown);
      if (state != BluetoothAdapterState.on) {
        _addLog('[ERROR] Bluetooth is not on (state: $state)');
        setState(() => _scanning = false);
        return;
      }

      final transport = BleTransport();
      final devices = await transport.scan();
      setState(() {
        _devices = devices;
        _scanning = false;
      });
      _addLog('Found ${devices.length} device(s)');
    } catch (e) {
      _addLog('[ERROR] Scan failed: $e');
      setState(() => _scanning = false);
    }
  }

  Future<void> _runTests(ScannedDevice device) async {
    if (_running) return;
    setState(() {
      _running = true;
      _logs.clear();
    });

    await _testRunner.runAll(device: device);
    setState(() => _running = false);
  }

  Color _logColor(String line) {
    if (line.startsWith('[PASS]')) return BleColors.success;
    if (line.startsWith('[FAIL]') || line.startsWith('[ERROR]')) {
      return BleColors.error;
    }
    if (line.startsWith('[BENCH]')) return BleColors.accent;
    return BleColors.text;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const BleWordmark(fontSize: 22),
      ),
      body: Padding(
        padding: const EdgeInsets.all(BleSpacing.s4),
        child: Column(
          children: [
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _scanning || _running ? null : _scan,
                child: _scanning
                    ? const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: BleColors.textSecondary,
                            ),
                          ),
                          SizedBox(width: BleSpacing.s2),
                          Text('Scanning…'),
                        ],
                      )
                    : const Text('Scan'),
              ),
            ),
            if (_devices.isNotEmpty) ...[
              const SizedBox(height: BleSpacing.s3),
              Row(
                children: [
                  const Text(
                    'DEVICES',
                    style: TextStyle(
                      color: BleColors.textSecondary,
                      fontWeight: FontWeight.w600,
                      fontSize: 12,
                      letterSpacing: 0.8,
                    ),
                  ),
                  const SizedBox(width: BleSpacing.s2),
                  BleBadge('${_devices.length}'),
                ],
              ),
              const SizedBox(height: BleSpacing.s2),
              Container(
                constraints: const BoxConstraints(maxHeight: 220),
                decoration: BoxDecoration(
                  color: BleColors.bgSecondary,
                  border: Border.all(color: BleColors.border),
                  borderRadius: BorderRadius.circular(BleRadius.xl),
                ),
                clipBehavior: Clip.antiAlias,
                child: ListView.separated(
                  shrinkWrap: true,
                  itemCount: _devices.length,
                  separatorBuilder: (_, __) =>
                      const Divider(height: 1, color: BleColors.border),
                  itemBuilder: (context, index) {
                    final d = _devices[index];
                    return InkWell(
                      onTap: _running ? null : () => _runTests(d),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                            horizontal: BleSpacing.s3, vertical: 10),
                        child: Row(
                          children: [
                            BleSignalBars(level: rssiToLevel(d.rssi)),
                            const SizedBox(width: BleSpacing.s3),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    d.name ?? 'Unknown',
                                    style: const TextStyle(
                                      color: BleColors.text,
                                      fontSize: 15,
                                      fontWeight: FontWeight.w500,
                                    ),
                                  ),
                                  Text(
                                    d.address,
                                    style: const TextStyle(
                                      color: BleColors.textSecondary,
                                      fontSize: 11,
                                      fontFamily: 'monospace',
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(width: BleSpacing.s2),
                            Text(
                              '${d.rssi} dBm',
                              style: const TextStyle(
                                color: BleColors.textSecondary,
                                fontSize: 13,
                                fontFamily: 'monospace',
                              ),
                            ),
                            const Icon(
                              Icons.chevron_right,
                              size: 18,
                              color: BleColors.textSecondary,
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
            const SizedBox(height: BleSpacing.s3),
            Row(
              children: [
                const Spacer(),
                TextButton.icon(
                  onPressed: _logs.isEmpty
                      ? null
                      : () {
                          Clipboard.setData(
                              ClipboardData(text: _logs.join('\n')));
                          setState(() => _showCopied = true);
                          Future.delayed(const Duration(milliseconds: 1500),
                              () {
                            if (mounted) setState(() => _showCopied = false);
                          });
                        },
                  icon: Icon(
                    _showCopied ? Icons.check : Icons.copy,
                    size: 16,
                    color: _logs.isEmpty
                        ? BleColors.textSecondary
                        : BleColors.accent,
                  ),
                  label: Text(
                    _showCopied ? 'Copied!' : 'Copy Logs',
                    style: TextStyle(
                      fontSize: 13,
                      color: _logs.isEmpty
                          ? BleColors.textSecondary
                          : BleColors.accent,
                    ),
                  ),
                ),
              ],
            ),
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: BleColors.bgCode,
                  border: Border.all(color: BleColors.border),
                  borderRadius: BorderRadius.circular(BleRadius.lg),
                ),
                child: _logs.isEmpty
                    ? const Center(
                        child: Text(
                          'Scan for devices, then tap one to run tests.',
                          style: TextStyle(color: BleColors.textSecondary),
                        ),
                      )
                    : ListView.builder(
                        controller: _scrollController,
                        padding: const EdgeInsets.all(BleSpacing.s3),
                        itemCount: _logs.length,
                        itemBuilder: (context, index) {
                          final line = _logs[index];
                          return Padding(
                            padding: const EdgeInsets.symmetric(vertical: 1),
                            child: Text(
                              line,
                              style: TextStyle(
                                fontFamily: 'monospace',
                                fontSize: 13,
                                color: _logColor(line),
                              ),
                            ),
                          );
                        },
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }
}
