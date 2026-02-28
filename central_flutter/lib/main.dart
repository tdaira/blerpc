import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
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
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(12.0),
            child: Row(
              children: [
                ElevatedButton.icon(
                  onPressed: _scanning || _running ? null : _scan,
                  icon: _scanning
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(
                              strokeWidth: 2, color: Colors.white))
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
            child: Container(
              margin: const EdgeInsets.all(12),
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
                              fontSize: 12,
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
    );
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }
}
