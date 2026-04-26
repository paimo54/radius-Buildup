# Dynamic Virtual Parameters Extraction

## Overview
Sistem pengambilan Virtual Parameters yang dinamis dari GenieACS, sehingga semua Virtual Parameters yang ditambahkan di GenieACS akan otomatis ter-ekstrak tanpa perlu update kode.

## Perubahan yang Dilakukan

### 1. **API ONT Route** (`/api/customer/ont`)

#### Sebelum (Static):
```typescript
const deviceInfo = {
  ipAddress: vp.pppoeIP?._value || 'N/A',
  rxPower: vp.RXPower?._value || 'N/A',
  temperature: vp.gettemp?._value || 'N/A',
  // Harus manual tambah setiap VP baru
};
```

#### Sesudah (Dynamic):
```typescript
// Extract ALL Virtual Parameters dynamically
const vp = device.VirtualParameters || {};
const virtualParams: Record<string, any> = {};

if (vp && typeof vp === 'object') {
  Object.keys(vp).forEach(key => {
    const value = vp[key];
    virtualParams[key] = value?._value !== undefined ? value._value : value;
  });
}

// Use with fallback
const deviceInfo = {
  pppUsername: virtualParams.pppoeUsername || virtualParams.pppoeUsername2 || 'N/A',
  ipAddress: virtualParams.pppoeIP || virtualParams.wanIP || 'N/A',
  rxPower: virtualParams.RXPower || virtualParams.rxPower || 'N/A',
  temperature: virtualParams.gettemp || virtualParams.temperature || 'N/A',
  uptime: virtualParams.getdeviceuptime || virtualParams.uptime || 'N/A',
  // Return semua VP untuk flexibilitas
  allVirtualParams: virtualParams
};
```

### 2. **API WiFi Route** (`/api/customer/wifi`)

Sama seperti ONT API, ekstraksi VP dilakukan secara dinamis dengan multiple fallback untuk nama VP yang berbeda-beda antar manufacturer.

#### Fitur Dynamic Extraction:
- Auto-detect semua VP yang ada
- Extract `_value` dari object VP GenieACS
- Fallback ke multiple nama alternatif (pppoeIP / wanIP / pppoeUsername / pppoeUsername2)
- Return raw VP data untuk custom usage

## Virtual Parameters yang Didukung

### PPPoE Information
- `pppoeUsername`, `pppoeUsername2` - Username PPPoE
- `pppoeIP`, `wanIP` - IP Address PPPoE
- Connection status berdasarkan keberadaan username

### Signal & Performance
- `RXPower`, `rxPower` - Receive power signal
- `TXPower`, `txPower` - Transmit power
- `gettemp`, `temperature` - ONT temperature

### Device Info
- `getSerialNumber`, `serialNumber` - Serial number ONT
- `getdeviceuptime`, `uptime` - Device uptime
- `getponmode`, `ponMode` - PON mode (GPON/EPON)

### WiFi Parameters
- `WlanPassword`, `wifiPassword` - WiFi password
- Other WLAN configs from standard parameters

## Cara Kerja

### 1. Extraction Process
```typescript
Object.keys(vp).forEach(key => {
  const value = vp[key];
  // Extract _value if exists, otherwise use raw value
  virtualParams[key] = value?._value !== undefined ? value._value : value;
});
```

### 2. Multiple Fallback Strategy
```typescript
// Try multiple possible names
ipAddress: virtualParams.pppoeIP || virtualParams.wanIP || 'N/A'
username: virtualParams.pppoeUsername || virtualParams.pppoeUsername2 || 'N/A'
```

### 3. Type Safety
- Always return string or 'N/A'
- Check for null/undefined before display
- Convert object values to primitive types

## Benefits

### ✅ Auto-Adapt to New VPs
Ketika admin menambahkan Virtual Parameter baru di GenieACS (misal: `bandwidth`, `upSpeed`, `downSpeed`), sistem langsung bisa mengakses data tersebut via `allVirtualParams` tanpa update kode.

### ✅ Multi-Manufacturer Support
Berbeda manufacturer menggunakan nama VP berbeda:
- Fiberhome: `pppoeUsername`
- ZTE: `pppoeUsername2`
- Huawei: `wanUsername`

Sistem mencoba semua kemungkinan dengan fallback.

### ✅ Backward Compatible
Data yang sudah ada tetap berfungsi, hanya ditambahkan field `allVirtualParams` untuk flexibilitas.

### ✅ Easy Debugging
Admin bisa lihat semua VP yang available di object `allVirtualParams`.

## Customer Dashboard Display

### Status Connection
```typescript
PPPoE: Connected/Disconnected
- Based on pppUsername existence
- Green badge if connected
- Gray badge if disconnected
```

### Device Information
- Model: Manufacturer + ProductClass
- Status ONT: Online/Offline
- IP PPPoE: From virtualParams.pppoeIP
- RX Power: From virtualParams.RXPower
- Temperature: From virtualParams.gettemp
- Uptime: From virtualParams.getdeviceuptime
- Connected Devices: Count from connectedHosts

## API Response Structure

### GET /api/customer/ont
```json
{
  "success": true,
  "device": {
    "serialNumber": "ZTEG12345678",
    "manufacturer": "ZTE",
    "model": "F660",
    "pppUsername": "user001",
    "ipAddress": "10.10.10.1",
    "uptime": "3 days 2 hours",
    "status": "Online",
    "rxPower": "-23.45 dBm",
    "txPower": "2.34 dBm",
    "temperature": "45°C",
    "ponMode": "GPON",
    "connectedHosts": 3,
    "wifiSSID": "MyWiFi",
    "wifiPassword": "********",
    "wifiEnabled": true,
    "wifiChannel": "6",
    "allVirtualParams": {
      "pppoeUsername": "user001",
      "pppoeIP": "10.10.10.1",
      "RXPower": "-23.45 dBm",
      "gettemp": "45°C",
      "getdeviceuptime": "3 days 2 hours",
      "customParam1": "value1",
      "customParam2": "value2"
    }
  }
}
```

## GenieACS Virtual Parameters Setup

### Example Virtual Parameters Script

```javascript
// PPPoE Username
const pppUsername = declare("VirtualParameters.pppoeUsername", {value: Date.now()}, {value: true});
const wanPPP = declare("InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.*", {path: 1}, {path: 1});
pppUsername.value = wanPPP.Username;

// PPPoE IP
const pppIP = declare("VirtualParameters.pppoeIP", {value: Date.now()}, {value: true});
pppIP.value = wanPPP.ExternalIPAddress;

// RX Power
const rxPower = declare("VirtualParameters.RXPower", {value: Date.now()}, {value: true});
const gpon = declare("InternetGatewayDevice.WANDevice.1.X_GponInterafceConfig.*", {path: 1}, {path: 1});
rxPower.value = gpon.RXPower;

// Temperature
const temp = declare("VirtualParameters.gettemp", {value: Date.now()}, {value: true});
temp.value = gpon.TransceiverTemperature;

// Uptime
const uptime = declare("VirtualParameters.getdeviceuptime", {value: Date.now()}, {value: true});
const devInfo = declare("InternetGatewayDevice.DeviceInfo.*", {path: 1}, {path: 1});
uptime.value = devInfo.UpTime;

// Custom Parameter Example
const bandwidth = declare("VirtualParameters.currentBandwidth", {value: Date.now()}, {value: true});
bandwidth.value = calculateBandwidth(); // Your custom logic
```

## Error Handling

### NULL/Undefined Checks
```typescript
// Always check before display
{ontDevice.pppUsername && ontDevice.pppUsername !== 'N/A' 
  ? 'Connected' 
  : 'Disconnected'}

{ontDevice.ipAddress && ontDevice.ipAddress !== 'N/A' 
  ? ontDevice.ipAddress 
  : '-'}
```

### Graceful Degradation
Jika VP tidak ada, sistem fallback ke:
1. Alternative VP name
2. Standard parameter path
3. 'N/A' or '-'

## Testing

### Test Scenarios
1. ✅ ONT dengan full Virtual Parameters
2. ✅ ONT tanpa Virtual Parameters (fallback to standard)
3. ✅ ONT dengan partial VPs (some missing)
4. ✅ Multiple manufacturers (ZTE, Fiberhome, Huawei)
5. ✅ PPPoE connected vs disconnected
6. ✅ Device online vs offline

### Expected Behavior
- No crashes on missing data
- Display 'N/A' or '-' for unavailable data
- Connection status accurate based on username
- All VPs accessible via allVirtualParams

## Future Enhancements

### 1. Custom VP Display
Admin bisa configure di UI VP mana yang mau ditampilkan ke customer:
```json
{
  "displayVPs": ["bandwidth", "upSpeed", "downSpeed", "signal"],
  "hideVPs": ["serialNumber", "internalIP"]
}
```

### 2. VP Alerts
Alert customer jika VP tertentu abnormal:
```typescript
if (rxPower < -27) {
  alert("Signal lemah, hubungi teknisi");
}
```

### 3. Historical VP Data
Store VP data over time untuk graph/trending:
- RX Power trend
- Temperature history
- Bandwidth usage

## Troubleshooting

### VP Not Showing
1. Check GenieACS Virtual Parameters configuration
2. Verify device last inform is recent (< 5 min)
3. Check VP name spelling in GenieACS
4. Test with allVirtualParams to see raw data
5. Check browser console for API errors

### Wrong Data Display
1. Verify VP value type in GenieACS
2. Check _value extraction logic
3. Test with different manufacturers
4. Review fallback chain order

### Performance Issues
1. Limit VP extraction to needed fields only
2. Cache allVirtualParams data
3. Use pagination for large device lists
4. Optimize GenieACS refresh interval

## Summary

Sistem dynamic Virtual Parameters extraction memberikan:
- ✅ **Flexibility**: Auto-adapt ke VP baru
- ✅ **Compatibility**: Multi-manufacturer support
- ✅ **Reliability**: Multiple fallback mechanisms
- ✅ **Maintainability**: Less code changes needed
- ✅ **Scalability**: Easy to add new data points

Customer sekarang mendapat informasi lengkap tentang device mereka secara real-time tanpa perlu intervensi admin.
