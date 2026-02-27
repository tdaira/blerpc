import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';

import 'ble/ble_transport.dart';
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
      title: 'blerpc Central',
      theme: ThemeData(
        colorSchemeSeed: Colors.blue,
        useMaterial3: true,
      ),
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
      final state = await FlutterBluePlus.adapterState.first;
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('blerpc Central')),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(8.0),
            child: Row(
              children: [
                ElevatedButton.icon(
                  onPressed: _scanning || _running ? null : _scan,
                  icon: _scanning
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2))
                      : const Icon(Icons.bluetooth_searching),
                  label: const Text('Scan'),
                ),
                const SizedBox(width: 8),
                if (_devices.isNotEmpty)
                  Expanded(
                    child: SizedBox(
                      height: 48,
                      child: ListView.separated(
                        scrollDirection: Axis.horizontal,
                        itemCount: _devices.length,
                        separatorBuilder: (_, __) => const SizedBox(width: 4),
                        itemBuilder: (context, index) {
                          final d = _devices[index];
                          return ActionChip(
                            label: Text(d.name ?? d.address),
                            onPressed: _running ? null : () => _runTests(d),
                          );
                        },
                      ),
                    ),
                  ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: _logs.isEmpty
                ? const Center(
                    child: Text('Scan for devices, then tap to run tests'))
                : ListView.builder(
                    controller: _scrollController,
                    padding: const EdgeInsets.all(8),
                    itemCount: _logs.length,
                    itemBuilder: (context, index) {
                      final line = _logs[index];
                      Color? color;
                      if (line.startsWith('[PASS]')) {
                        color = Colors.green;
                      } else if (line.startsWith('[FAIL]') ||
                          line.startsWith('[ERROR]')) {
                        color = Colors.red;
                      } else if (line.startsWith('[BENCH]')) {
                        color = Colors.blue;
                      }
                      return Text(
                        line,
                        style: TextStyle(
                          fontFamily: 'monospace',
                          fontSize: 12,
                          color: color,
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }
}
