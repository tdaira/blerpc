import React, { useState, useRef, useCallback } from 'react';
import {
  StyleSheet,
  View,
  Text,
  TouchableOpacity,
  FlatList,
  Platform,
  PermissionsAndroid,
} from 'react-native';
import Clipboard from '@react-native-clipboard/clipboard';
import { ScannedDevice } from './ble/BleTransport';
import { BlerpcClient } from './client/BlerpcClient';
import { TestRunner } from './test/TestRunner';

// blerpc.net dark theme colors
const colors = {
  bgPrimary: '#1A1B26',
  bgSecondary: '#24283B',
  bgCode: '#1E2030',
  textPrimary: '#C0CAF5',
  textSecondary: '#A9B1D6',
  accent: '#0082FC',
  border: '#3B4261',
  success: '#9ECE6A',
  error: '#F7768E',
  navBg: '#16161E',
};

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
    if (text.startsWith('[PASS]')) return colors.success;
    if (text.startsWith('[FAIL]') || text.startsWith('[ERROR]')) return colors.error;
    if (text.startsWith('[BENCH]')) return colors.accent;
    return colors.textPrimary;
  };

  return (
    <View style={styles.container}>
      <View style={styles.appBar}>
        <Text style={styles.appBarTitle}>
          <Text style={styles.titleBle}>ble</Text>
          <Text style={styles.titleRpc}>RPC</Text>
          <Text style={styles.titleCentral}> Central</Text>
        </Text>
      </View>

      <View style={styles.body}>
        <View style={styles.buttonRow}>
          <TouchableOpacity
            style={[styles.button, (scanning || running) && styles.buttonDisabled]}
            onPress={handleScan}
            disabled={scanning || running}
          >
            <Text style={styles.buttonText}>{scanning ? 'Scanning...' : 'Scan'}</Text>
          </TouchableOpacity>
        </View>

        {devices.length > 0 && (
          <View style={styles.deviceSection}>
            <Text style={styles.sectionTitle}>Devices ({devices.length})</Text>
            <View style={styles.deviceList}>
              {devices.map((d, i) => (
                <TouchableOpacity
                  key={d.address + i}
                  style={styles.deviceItem}
                  onPress={() => handleRunTests(d)}
                  disabled={running}
                >
                  <Text style={styles.deviceName}>{d.name ?? 'Unknown'}</Text>
                  <Text style={styles.deviceAddress}>{d.address}</Text>
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
              <Text style={styles.placeholderText}>Scan for devices, then tap to run tests</Text>
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
    backgroundColor: colors.bgPrimary,
  },
  appBar: {
    backgroundColor: colors.navBg,
    paddingTop: Platform.OS === 'ios' ? 54 : 40,
    paddingBottom: 12,
    paddingHorizontal: 16,
  },
  appBarTitle: {
    fontSize: 20,
  },
  titleBle: {
    color: colors.accent,
    fontWeight: '900',
  },
  titleRpc: {
    color: colors.textPrimary,
    fontWeight: '900',
  },
  titleCentral: {
    color: colors.textPrimary,
    fontWeight: '400',
  },
  body: {
    flex: 1,
    padding: 16,
  },
  buttonRow: {
    flexDirection: 'row',
    gap: 12,
  },
  button: {
    flex: 1,
    backgroundColor: colors.accent,
    borderRadius: 8,
    paddingVertical: 12,
    alignItems: 'center',
  },
  buttonDisabled: {
    backgroundColor: colors.bgSecondary,
  },
  buttonText: {
    color: '#FFFFFF',
    fontWeight: '600',
    fontSize: 16,
  },
  deviceSection: {
    marginTop: 12,
  },
  sectionTitle: {
    color: colors.textPrimary,
    fontWeight: '600',
    fontSize: 14,
    marginBottom: 4,
  },
  deviceList: {
    backgroundColor: colors.bgSecondary,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 8,
    maxHeight: 200,
  },
  deviceItem: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
  },
  deviceName: {
    color: colors.textPrimary,
    fontSize: 15,
    fontWeight: '500',
  },
  deviceAddress: {
    color: colors.textSecondary,
    fontSize: 11,
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
  },
  logHeader: {
    flexDirection: 'row',
    marginTop: 12,
    marginBottom: 4,
  },
  spacer: {
    flex: 1,
  },
  copyText: {
    color: colors.accent,
    fontSize: 13,
  },
  copyTextDisabled: {
    color: colors.textSecondary,
  },
  logContainer: {
    flex: 1,
    backgroundColor: colors.bgCode,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 8,
    overflow: 'hidden',
  },
  logPlaceholder: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  placeholderText: {
    color: colors.textSecondary,
  },
  logContent: {
    padding: 12,
  },
  logLine: {
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
    fontSize: 13,
    lineHeight: 18,
  },
});
