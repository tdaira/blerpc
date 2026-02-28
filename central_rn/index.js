import 'fast-text-encoding';
import 'react-native-get-random-values';
import { AppRegistry, Alert } from 'react-native';
import App from './src/App';

// Global error handler to prevent silent crashes in release mode
const defaultHandler = ErrorUtils.getGlobalHandler();
ErrorUtils.setGlobalHandler((error, isFatal) => {
  Alert.alert(
    isFatal ? 'Fatal Error' : 'Error',
    String(error?.message || error),
  );
  if (defaultHandler) defaultHandler(error, isFatal);
});

AppRegistry.registerComponent('BlerpcCentral', () => App);
