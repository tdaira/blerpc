import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter/services.dart';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';

import 'ble/ble_transport.dart';
import 'test/test_runner.dart';

const _autoRun = bool.fromEnvironment('AUTO_RUN', defaultValue: false);

// blerpc.net dark theme colors
const _bgPrimary = Color(0xFF1A1B26);
const _bgSecondary = Color(0xFF24283B);
const _bgCode = Color(0xFF1E2030);
const _textPrimary = Color(0xFFC0CAF5);
const _textSecondary = Color(0xFFA9B1D6);
const _accent = Color(0xFF0082FC);
const _border = Color(0xFF3B4261);
const _success = Color(0xFF9ECE6A);
const _error = Color(0xFFF7768E);
const _navBg = Color(0xFF16161E);

void main() {
  runApp(const BlerpcCentralApp());
}

class BlerpcCentralApp extends StatelessWidget {
  const BlerpcCentralApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'bleRPC Central',
      theme: ThemeData(
        brightness: Brightness.dark,
        scaffoldBackgroundColor: _bgPrimary,
        colorScheme: const ColorScheme.dark(
          primary: _accent,
          onPrimary: Colors.white,
          surface: _bgPrimary,
          onSurface: _textPrimary,
          secondary: _accent,
          outline: _border,
        ),
        appBarTheme: const AppBarTheme(
          backgroundColor: _navBg,
          foregroundColor: _textPrimary,
          elevation: 0,
        ),
        dividerTheme: const DividerThemeData(color: _border),
        chipTheme: ChipThemeData(
          backgroundColor: _bgSecondary,
          labelStyle: const TextStyle(color: _textPrimary),
          side: const BorderSide(color: _border),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: _accent,
            foregroundColor: Colors.white,
            disabledBackgroundColor: _bgSecondary,
            disabledForegroundColor: _textSecondary,
          ),
        ),
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text.rich(TextSpan(children: [
          const TextSpan(
              text: 'ble',
              style: TextStyle(color: _accent, fontWeight: FontWeight.w900)),
          TextSpan(
              text: 'RPC',
              style:
                  TextStyle(color: _textPrimary, fontWeight: FontWeight.w900)),
          const TextSpan(
              text: ' Central', style: TextStyle(fontWeight: FontWeight.w400)),
        ])),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Row(
              children: [
                Expanded(
                  child: ElevatedButton(
                    onPressed: _scanning || _running ? null : _scan,
                    child: Text(_scanning ? 'Scanning...' : 'Scan'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton(
                    onPressed: _running || _scanning ? null : null,
                    child: Text(_running ? 'Running...' : 'Run Tests'),
                  ),
                ),
              ],
            ),
            if (_devices.isNotEmpty) ...[
              const SizedBox(height: 12),
              Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Devices (${_devices.length})',
                  style: const TextStyle(
                    color: _textPrimary,
                    fontWeight: FontWeight.w600,
                    fontSize: 14,
                  ),
                ),
              ),
              const SizedBox(height: 4),
              Container(
                constraints: const BoxConstraints(maxHeight: 200),
                decoration: BoxDecoration(
                  color: _bgSecondary,
                  border: Border.all(color: _border),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: ListView.separated(
                  shrinkWrap: true,
                  itemCount: _devices.length,
                  separatorBuilder: (_, __) =>
                      const Divider(height: 1, color: _border),
                  itemBuilder: (context, index) {
                    final d = _devices[index];
                    return InkWell(
                      onTap: _running ? null : () => _runTests(d),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 16, vertical: 8),
                        child: Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    d.name ?? 'Unknown',
                                    style: const TextStyle(
                                      color: _textPrimary,
                                      fontSize: 15,
                                      fontWeight: FontWeight.w500,
                                    ),
                                  ),
                                  Text(
                                    d.address,
                                    style: const TextStyle(
                                      color: _textSecondary,
                                      fontSize: 11,
                                      fontFamily: 'monospace',
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
            const SizedBox(height: 12),
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
                    color: _logs.isEmpty ? _textSecondary : _accent,
                  ),
                  label: Text(
                    _showCopied ? 'Copied!' : 'Copy Logs',
                    style: TextStyle(
                      fontSize: 13,
                      color: _logs.isEmpty ? _textSecondary : _accent,
                    ),
                  ),
                ),
              ],
            ),
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: _bgCode,
                  border: Border.all(color: _border),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: _logs.isEmpty
                    ? const Center(
                        child: Text(
                          'Scan for devices, then tap to run tests',
                          style: TextStyle(color: _textSecondary),
                        ),
                      )
                    : ListView.builder(
                        controller: _scrollController,
                        padding: const EdgeInsets.all(12),
                        itemCount: _logs.length,
                        itemBuilder: (context, index) {
                          final line = _logs[index];
                          Color color = _textPrimary;
                          if (line.startsWith('[PASS]')) {
                            color = _success;
                          } else if (line.startsWith('[FAIL]') ||
                              line.startsWith('[ERROR]')) {
                            color = _error;
                          } else if (line.startsWith('[BENCH]')) {
                            color = _accent;
                          }
                          return Padding(
                            padding: const EdgeInsets.symmetric(vertical: 1),
                            child: Text(
                              line,
                              style: TextStyle(
                                fontFamily: 'monospace',
                                fontSize: 13,
                                color: color,
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
