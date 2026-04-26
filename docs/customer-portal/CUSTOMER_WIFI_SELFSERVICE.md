# Customer WiFi Self-Service Documentation

## 📋 Overview

Fitur WiFi Self-Service memungkinkan pelanggan untuk mengelola konfigurasi WiFi mereka sendiri tanpa perlu menghubungi admin. Sistem ini terintegrasi dengan GenieACS untuk manajemen perangkat ONT/CPE secara real-time.

## ✨ Features

### 1. WiFi Information Dashboard
- **View WiFi Networks**: Melihat semua jaringan WiFi (2.4GHz dan 5GHz)
- **Device Status**: Status online/offline perangkat ONT
- **Connected Devices**: Jumlah device yang terhubung ke WiFi
- **Network Details**: SSID, Channel, Security, Standard (802.11n/ac/ax)
- **Signal Information**: RX Power, TX Power, Temperature

### 2. WiFi Management
- **Edit SSID**: Mengubah nama WiFi (1-32 karakter)
- **Change Password**: Mengubah password WiFi (8-63 karakter)
- **Security Mode**: Support WPA2-PSK encryption
- **Auto Reboot**: Device otomatis restart setelah perubahan
- **Real-time Update**: Perubahan langsung apply ke perangkat

### 3. Connected Devices Management
- **Device List**: Daftar semua device yang terhubung
- **Device Info**: MAC Address, IP Address, Hostname
- **Device Type Detection**: Auto-detect Android, iPhone, Laptop, Desktop
- **Active Status**: Status aktif/idle setiap device
- **Signal Strength**: Kekuatan sinyal WiFi per device

### 4. Package Upgrade via Payment Gateway
- **View Packages**: Lihat semua paket internet available
- **Compare Plans**: Bandingkan kecepatan dan harga
- **Upgrade Request**: Request upgrade paket
- **Auto Invoice**: Generate invoice otomatis
- **Payment Gateway**: Redirect ke payment gateway (Midtrans)
- **Auto Apply**: Paket otomatis aktif setelah pembayaran

## 🎨 User Interface

### WiFi Dashboard
```
┌─────────────────────────────────────────┐
│  Manajemen WiFi              [Refresh]  │
├─────────────────────────────────────────┤
│  📊 Status Perangkat                    │
│  Model: Huawei HG8145V5                 │
│  Serial: 48575443C7XXXXXX               │
│  Status: ● Online                       │
├─────────────────────────────────────────┤
│  📡 WiFi 2.4GHz          🔧 [Edit]     │
│  SSID: MyWiFi                           │
│  Band: 2.4GHz | Channel: 6              │
│  Security: WPA2-PSK | Devices: 3        │
│─────────────────────────────────────────│
│  📡 WiFi 5GHz            🔧 [Edit]     │
│  SSID: MyWiFi-5G                        │
│  Band: 5GHz | Channel: 149              │
│  Security: WPA2-PSK | Devices: 2        │
├─────────────────────────────────────────┤
│  📱 Perangkat Terhubung (5)             │
│  ┌────────────────────────────────────┐ │
│  │ 📱 Samsung Galaxy | 192.168.1.10   │ │
│  │ 💻 MacBook Pro    | 192.168.1.11   │ │
│  │ 📱 iPhone 13      | 192.168.1.12   │ │
│  └────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│  ⚡ Upgrade Paket Internet              │
│  [10Mbps] [20Mbps] [50Mbps]            │
│  [Lanjutkan ke Pembayaran]             │
└─────────────────────────────────────────┘
```

### WiFi Edit Form
```
┌─────────────────────────────────────────┐
│  Edit WiFi Configuration                │
├─────────────────────────────────────────┤
│  SSID Baru                              │
│  [MyNewWiFi____________]                │
│                                         │
│  Password Baru                          │
│  [••••••••••] 👁                        │
│                                         │
│  ⚠️ Perangkat akan restart.             │
│  Koneksi WiFi akan terputus sementara.  │
│                                         │
│  [Simpan Perubahan] [Batal]            │
└─────────────────────────────────────────┘
```

## 🔧 Technical Implementation

### Frontend
**File**: `src/app/customer/wifi/page.tsx`

**Features**:
- React hooks for state management
- Real-time data fetching
- Form validation
- SweetAlert2 for confirmations
- Responsive design (mobile-first)
- Dark mode support

**Key Components**:
```tsx
- WiFi Information Card
- WiFi Edit Form
- Connected Devices Table
- Package Selection Grid
- Device Status Indicators
```

### Backend APIs

#### 1. Get WiFi Info
**Endpoint**: `GET /api/customer/wifi`

**Headers**:
```json
{
  "Authorization": "Bearer <customer_token>"
}
```

**Response**:
```json
{
  "success": true,
  "device": {
    "_id": "00-50-56-XX-XX-XX",
    "pppUsername": "customer1@realm",
    "serialNumber": "48575443C7XXXXXX",
    "model": "EchoLife HG8145V5",
    "manufacturer": "Huawei",
    "status": "Online",
    "wlanConfigs": [
      {
        "index": 1,
        "ssid": "MyWiFi",
        "enabled": true,
        "channel": "6",
        "standard": "802.11n",
        "security": "WPA2-PSK",
        "password": "********",
        "band": "2.4GHz",
        "totalAssociations": 3,
        "bssid": "XX:XX:XX:XX:XX:XX"
      }
    ],
    "connectedHosts": [
      {
        "macAddress": "XX:XX:XX:XX:XX:XX",
        "ipAddress": "192.168.1.10",
        "hostname": "Samsung-Galaxy",
        "associatedDevice": "MyWiFi",
        "active": true,
        "signalStrength": "-45 dBm"
      }
    ]
  }
}
```

#### 2. Update WiFi
**Endpoint**: `POST /api/customer/wifi`

**Headers**:
```json
{
  "Authorization": "Bearer <customer_token>",
  "Content-Type": "application/json"
}
```

**Request Body**:
```json
{
  "deviceId": "00-50-56-XX-XX-XX",
  "wlanIndex": 1,
  "ssid": "MyNewWiFi",
  "password": "NewPassword123",
  "securityMode": "WPA2-PSK",
  "enabled": true
}
```

**Response**:
```json
{
  "success": true,
  "message": "WiFi configuration updated successfully"
}
```

#### 3. Get Packages
**Endpoint**: `GET /api/customer/packages`

**Response**:
```json
{
  "success": true,
  "packages": [
    {
      "id": "pkg-001",
      "name": "Paket 10Mbps",
      "downloadSpeed": 10,
      "uploadSpeed": 10,
      "price": 150000,
      "description": "Paket hemat untuk browsing"
    }
  ]
}
```

#### 4. Upgrade Package
**Endpoint**: `POST /api/customer/upgrade-package`

**Request Body**:
```json
{
  "packageId": "pkg-002"
}
```

**Response**:
```json
{
  "success": true,
  "invoice": {
    "id": "inv-xxx",
    "invoiceNumber": "INV/2025/01/15/0001",
    "amount": 250000,
    "dueDate": "2025-01-22T00:00:00.000Z",
    "paymentLink": "https://app.midtrans.com/snap/v2/..."
  },
  "paymentLink": "https://app.midtrans.com/snap/v2/..."
}
```

## 🔌 GenieACS Integration

### TR-069 Parameters Used

#### WiFi Configuration
```
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.SSID
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.Enable
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.BeaconType
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.KeyPassphrase
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.PreSharedKey.1.KeyPassphrase
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.Channel
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.Standard
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.TotalAssociations
```

#### Device Information
```
InternetGatewayDevice.DeviceInfo.SerialNumber
InternetGatewayDevice.DeviceInfo.ModelName
InternetGatewayDevice.DeviceInfo.Manufacturer
InternetGatewayDevice.DeviceInfo.SoftwareVersion
```

#### Connected Devices
```
InternetGatewayDevice.LANDevice.1.Hosts.Host.{index}.MACAddress
InternetGatewayDevice.LANDevice.1.Hosts.Host.{index}.IPAddress
InternetGatewayDevice.LANDevice.1.Hosts.Host.{index}.HostName
InternetGatewayDevice.LANDevice.1.WLANConfiguration.{index}.AssociatedDevice.{index}
```

### GenieACS Tasks

#### Set WiFi Parameters
```javascript
{
  "name": "setParameterValues",
  "parameterValues": [
    ["InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.SSID", "NewSSID", "xsd:string"],
    ["InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.KeyPassphrase", "NewPassword", "xsd:string"]
  ]
}
```

#### Reboot Device
```javascript
{
  "name": "reboot"
}
```

## 🔐 Security Features

### Customer Authentication
- JWT token-based authentication
- Token expiry (24 hours default)
- Secure password hashing (bcrypt)
- IP address logging

### WiFi Security
- WPA2-PSK encryption
- Password complexity validation (8-63 characters)
- SSID length validation (1-32 characters)
- XSS protection on inputs

### API Security
- Rate limiting per customer
- CORS protection
- Request validation
- Error sanitization

## 📱 Responsive Design

### Mobile (< 768px)
- Stacked WiFi cards
- Full-width forms
- Simplified device list
- Touch-optimized buttons

### Tablet (768px - 1024px)
- 2-column WiFi grid
- Compact device table
- Side-by-side package cards

### Desktop (> 1024px)
- 2-column WiFi grid
- Full device table with all columns
- 3-column package grid
- Enhanced hover effects

## 🎨 UI/UX Best Practices

### Colors
- **Primary**: Teal (#0d9488) - Actions, WiFi 2.4GHz
- **Secondary**: Purple (#9333ea) - WiFi 5GHz
- **Success**: Green (#22c55e) - Online, Active
- **Warning**: Yellow (#eab308) - Alerts
- **Danger**: Red (#ef4444) - Offline, Error

### Icons
- 📡 Wifi - WiFi networks
- 📱 Smartphone - Mobile devices
- 💻 Laptop - Laptops
- 🖥️ Monitor - Desktops
- ⚡ Zap - Package upgrade
- 🔄 RefreshCw - Refresh data
- ✏️ Edit2 - Edit WiFi
- 💾 Save - Save changes

### Feedback
- **Loading**: Spinner animation
- **Success**: SweetAlert2 success dialog
- **Error**: SweetAlert2 error dialog with details
- **Confirmation**: SweetAlert2 confirmation before actions

## 🚀 Usage Flow

### Change WiFi Name/Password

1. Customer login ke customer portal
2. Navigate ke "WiFi Management" menu
3. Click "Edit" button pada WiFi card
4. Input SSID dan password baru
5. Click "Simpan Perubahan"
6. Confirm perubahan pada dialog
7. System kirim task ke GenieACS
8. Device restart otomatis
9. Customer re-connect dengan WiFi baru

### Upgrade Package

1. Customer login ke customer portal
2. Navigate ke "WiFi Management" menu
3. Scroll ke "Upgrade Paket Internet" section
4. Pilih paket yang diinginkan
5. Click "Lanjutkan ke Pembayaran"
6. Confirm upgrade pada dialog
7. System create invoice
8. Redirect ke payment gateway
9. Complete payment
10. Package otomatis aktif setelah payment

## ⚙️ Configuration

### Environment Variables
```env
# GenieACS Configuration (already configured)
GENIEACS_URL=http://localhost:7557
GENIEACS_USERNAME=admin
GENIEACS_PASSWORD=password

# JWT Secret
JWT_SECRET=your-secret-key-change-in-production

# Midtrans Payment Gateway (optional)
MIDTRANS_SERVER_KEY=your-server-key
MIDTRANS_CLIENT_KEY=your-client-key
MIDTRANS_IS_PRODUCTION=false

# App URL
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### GenieACS Setup

Ensure GenieACS is configured with:
1. Device provisioning enabled
2. TR-069 parameters configured
3. Virtual parameters for PPPoE username matching

### Database Schema

No additional schema changes required. Uses existing tables:
- `pppoe_users` - Customer data
- `internet_profiles` - Package data
- `invoices` - Upgrade invoices
- `activity_logs` - Customer actions
- `genieacs_settings` - GenieACS config

## 🐛 Troubleshooting

### Device Not Found
**Cause**: Device not in GenieACS or PPPoE username mismatch

**Solution**:
1. Check GenieACS device list
2. Verify Virtual Parameter `pppUsername` is set
3. Ensure device has contacted GenieACS recently

### WiFi Update Failed
**Cause**: GenieACS task timeout or device offline

**Solution**:
1. Check device status (must be Online)
2. Verify GenieACS connection
3. Check device supports TR-069 WiFi parameters
4. Review GenieACS logs

### Connected Devices Not Showing
**Cause**: Device doesn't support AssociatedDevice parameter

**Solution**:
1. Check device TR-069 capabilities
2. Use Hosts.Host parameter as fallback
3. Update device firmware if needed

### Payment Gateway Not Working
**Cause**: Midtrans not configured or keys invalid

**Solution**:
1. Configure Midtrans keys in Company settings
2. Verify Midtrans account is active
3. Check Midtrans server/client keys
4. Review Midtrans logs

## 📊 Metrics & Monitoring

### Customer Actions Logged
- `update_wifi` - WiFi configuration changed
- `package_upgrade_request` - Package upgrade requested
- `wifi_view` - Customer viewed WiFi page

### Success Metrics
- WiFi update success rate
- Package upgrade conversion rate
- Average time to change WiFi
- Customer self-service adoption

## 🔮 Future Enhancements

### Planned Features
1. **WiFi Scheduling**: Set WiFi on/off times
2. **Guest Network**: Create temporary guest WiFi
3. **Parental Control**: Block specific devices/websites
4. **Speed Test**: Built-in speedtest integration
5. **Usage Statistics**: WiFi usage per device
6. **QR Code**: Generate QR for WiFi sharing
7. **Multiple Devices**: Support multiple ONTs per customer
8. **Notification**: Push notification on WiFi changes

## 📚 References

- **GenieACS Documentation**: https://genieacs.com/docs/
- **TR-069 Protocol**: ITU-T Recommendation
- **gembok-bill Reference**: https://github.com/alijayanet/gembok-bill
- **Midtrans Snap**: https://docs.midtrans.com/

## 📞 Support

For technical support or feature requests:
- Email: support@yourcompany.com
- Documentation: /docs/CUSTOMER_WIFI_SELFSERVICE.md
- GitHub Issues: Create issue with label `customer-portal`

---

**Version**: 2.8.0  
**Last Updated**: December 24, 2025  
**Status**: ✅ Ready for Production
