import React, { useState, useRef, useCallback } from 'react';
import {
  StyleSheet,
  View,
  Text,
  TouchableOpacity,
  FlatList,
  Platform,
  PermissionsAndroid,
  ActivityIndicator,
} from 'react-native';
import Clipboard from '@react-native-clipboard/clipboard';
import { ScannedDevice } from './ble/BleTransport';
import { BlerpcClient } from './client/BlerpcClient';
import { TestRunner } from './test/TestRunner';
import { BleColors, BleSpacing, BleRadius, rssiToLevel } from './design/tokens';
import { Wordmark, Badge, SignalBars, monoFont } from './design/components';

interface LogEntry {
  id: number;
  text: string;
}

let logIdCounter = 0;

export default function App() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [devices, setDevices] = useState<ScannedDevice[]>([]);
  const [scanning, setScanning] = useState(false);
  const [running, setRunning] = useState(false);
  const flatListRef = useRef<FlatList>(null);

  const addLog = useCallback((message: string) => {
    console.log(message);
    setLogs((prev) => [...prev, { id: ++logIdCounter, text: message }]);
    setTimeout(() => {
      flatListRef.current?.scrollToEnd({ animated: true });
    }, 50);
  }, []);

  // Single BlerpcClient instance — react-native-ble-plx requires exactly
  // one BleManager for the lifetime of the app.
  const clientRef = useRef<BlerpcClient | null>(null);
  if (!clientRef.current) {
    clientRef.current = new BlerpcClient(true);
  }

  const testRunnerRef = useRef<TestRunner | null>(null);
  if (!testRunnerRef.current) {
    testRunnerRef.current = new TestRunner(addLog);
  }

  const requestPermissions = async () => {
    if (Platform.OS === 'android') {
      const apiLevel = Platform.Version;
      if (apiLevel >= 31) {
        const results = await PermissionsAndroid.requestMultiple([
          PermissionsAndroid.PERMISSIONS.BLUETOOTH_SCAN,
          PermissionsAndroid.PERMISSIONS.BLUETOOTH_CONNECT,
          PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
        ]);
        const allGranted = Object.values(results).every(
          (r) => r === PermissionsAndroid.RESULTS.GRANTED,
        );
        if (!allGranted) {
          addLog('[ERROR] BLE permissions not granted');
          return false;
        }
      } else {
        const result = await PermissionsAndroid.request(
          PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
        );
        if (result !== PermissionsAndroid.RESULTS.GRANTED) {
          addLog('[ERROR] Location permission not granted');
          return false;
        }
      }
    }
    return true;
  };

  const handleScan = async () => {
    setScanning(true);
    setDevices([]);
    setLogs([]);
    logIdCounter = 0;

    const granted = await requestPermissions();
    if (!granted) {
      setScanning(false);
      return;
    }

    addLog('Scanning...');
    try {
      const found = await clientRef.current!.scan();
      setDevices(found);
      addLog(`Found ${found.length} device(s)`);
    } catch (e) {
      addLog(`[ERROR] Scan failed: ${e}`);
    }
    setScanning(false);
  };

  const handleRunTests = async (device: ScannedDevice) => {
    if (running) return;
    setRunning(true);
    setLogs([]);
    logIdCounter = 0;

    try {
      await testRunnerRef.current!.runAll({ device, client: clientRef.current! });
    } catch (e) {
      addLog(`[ERROR] Uncaught: ${e}`);
    }
    setRunning(false);
  };

  const handleCopyLogs = () => {
    const text = logs.map((l) => l.text).join('\n');
    Clipboard.setString(text);
  };

  const getLogColor = (text: string): string => {
    if (text.startsWith('[PASS]')) return BleColors.success;
    if (text.startsWith('[FAIL]') || text.startsWith('[ERROR]')) return BleColors.error;
    if (text.startsWith('[BENCH]')) return BleColors.accent;
    return BleColors.text;
  };

  const disabled = scanning || running;

  return (
    <View style={styles.container}>
      <View style={styles.appBar}>
        <Wordmark fontSize={20} />
      </View>

      <View style={styles.body}>
        <TouchableOpacity
          style={[styles.button, disabled && styles.buttonDisabled]}
          onPress={handleScan}
          disabled={disabled}
        >
          {scanning ? (
            <View style={styles.buttonContent}>
              <ActivityIndicator size="small" color={BleColors.textSecondary} />
              <Text style={[styles.buttonText, styles.buttonTextDisabled]}>Scanning…</Text>
            </View>
          ) : (
            <Text style={styles.buttonText}>Scan</Text>
          )}
        </TouchableOpacity>

        {devices.length > 0 && (
          <View style={styles.deviceSection}>
            <View style={styles.sectionHeader}>
              <Text style={styles.sectionTitle}>DEVICES</Text>
              <Badge text={String(devices.length)} tone="blue" />
            </View>
            <View style={styles.deviceList}>
              {devices.map((d, i) => (
                <TouchableOpacity
                  key={d.address + i}
                  style={[styles.deviceItem, i < devices.length - 1 && styles.deviceItemBorder]}
                  onPress={() => handleRunTests(d)}
                  disabled={running}
                >
                  <SignalBars level={rssiToLevel(d.rssi)} />
                  <View style={styles.deviceInfo}>
                    <Text style={styles.deviceName}>{d.name ?? 'Unknown'}</Text>
                    <Text style={styles.deviceAddress}>{d.address}</Text>
                  </View>
                  <Text style={styles.deviceRssi}>{d.rssi} dBm</Text>
                  <Text style={styles.chevron}>›</Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>
        )}

        <View style={styles.logHeader}>
          <View style={styles.spacer} />
          <TouchableOpacity onPress={handleCopyLogs} disabled={logs.length === 0}>
            <Text style={[styles.copyText, logs.length === 0 && styles.copyTextDisabled]}>
              Copy Logs
            </Text>
          </TouchableOpacity>
        </View>

        <View style={styles.logContainer}>
          {logs.length === 0 ? (
            <View style={styles.logPlaceholder}>
              <Text style={styles.placeholderText}>
                Scan for devices, then tap one to run tests.
              </Text>
            </View>
          ) : (
            <FlatList
              ref={flatListRef}
              data={logs}
              keyExtractor={(item) => String(item.id)}
              contentContainerStyle={styles.logContent}
              renderItem={({ item }) => (
                <Text style={[styles.logLine, { color: getLogColor(item.text) }]}>{item.text}</Text>
              )}
            />
          )}
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: BleColors.bg,
  },
  appBar: {
    backgroundColor: BleColors.navBg,
    paddingTop: Platform.OS === 'ios' ? 54 : 40,
    paddingBottom: BleSpacing.s3,
    paddingHorizontal: BleSpacing.s4,
  },
  body: {
    flex: 1,
    padding: BleSpacing.s4,
  },
  button: {
    backgroundColor: BleColors.accent,
    borderRadius: BleRadius.md,
    paddingVertical: BleSpacing.s3,
    alignItems: 'center',
  },
  buttonDisabled: {
    backgroundColor: BleColors.bgSecondary,
  },
  buttonContent: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: BleSpacing.s2,
  },
  buttonText: {
    color: BleColors.onAccent,
    fontWeight: '600',
    fontSize: 16,
  },
  buttonTextDisabled: {
    color: BleColors.textSecondary,
  },
  deviceSection: {
    marginTop: BleSpacing.s3,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: BleSpacing.s2,
    marginBottom: BleSpacing.s2,
  },
  sectionTitle: {
    color: BleColors.textSecondary,
    fontWeight: '600',
    fontSize: 12,
    letterSpacing: 0.8,
  },
  deviceList: {
    backgroundColor: BleColors.bgSecondary,
    borderColor: BleColors.border,
    borderWidth: 1,
    borderRadius: BleRadius.xl,
    maxHeight: 220,
    overflow: 'hidden',
  },
  deviceItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: BleSpacing.s3,
    paddingHorizontal: BleSpacing.s3,
    paddingVertical: 10,
  },
  deviceItemBorder: {
    borderBottomColor: BleColors.border,
    borderBottomWidth: 1,
  },
  deviceInfo: {
    flex: 1,
  },
  deviceName: {
    color: BleColors.text,
    fontSize: 15,
    fontWeight: '500',
  },
  deviceAddress: {
    color: BleColors.textSecondary,
    fontSize: 11,
    fontFamily: monoFont,
  },
  deviceRssi: {
    color: BleColors.textSecondary,
    fontSize: 13,
    fontFamily: monoFont,
  },
  chevron: {
    color: BleColors.textSecondary,
    fontSize: 18,
  },
  logHeader: {
    flexDirection: 'row',
    marginTop: BleSpacing.s3,
    marginBottom: BleSpacing.s1,
  },
  spacer: {
    flex: 1,
  },
  copyText: {
    color: BleColors.accent,
    fontSize: 13,
  },
  copyTextDisabled: {
    color: BleColors.textSecondary,
  },
  logContainer: {
    flex: 1,
    backgroundColor: BleColors.bgCode,
    borderColor: BleColors.border,
    borderWidth: 1,
    borderRadius: BleRadius.lg,
    overflow: 'hidden',
  },
  logPlaceholder: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  placeholderText: {
    color: BleColors.textSecondary,
  },
  logContent: {
    padding: BleSpacing.s3,
  },
  logLine: {
    fontFamily: monoFont,
    fontSize: 13,
    lineHeight: 18,
  },
});
