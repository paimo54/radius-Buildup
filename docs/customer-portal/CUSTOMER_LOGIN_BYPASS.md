# Customer Login Bypass System

## 📋 Overview

Sistem login customer mendukung dua mode:
1. **Login dengan OTP** - Customer harus verifikasi kode OTP yang dikirim via WhatsApp
2. **Login tanpa OTP (Bypass)** - Customer dapat login langsung tanpa OTP

## 🔧 Konfigurasi

### Mengaktifkan/Menonaktifkan OTP

1. Login ke **Admin Panel**
2. Navigate ke **WhatsApp** > **Notifications**
3. Toggle **"Enable Customer OTP Login"**

**Status OTP ON (Default)**:
- ✅ Customer harus input kode OTP dari WhatsApp
- ✅ Lebih secure
- ✅ Validasi nomor WhatsApp aktif
- ❌ Bergantung pada layanan WhatsApp

**Status OTP OFF (Bypass Mode)**:
- ✅ Customer login langsung tanpa OTP
- ✅ Tidak bergantung WhatsApp
- ✅ Login lebih cepat
- ❌ Kurang secure (hanya validasi nomor terdaftar)

## 🎯 Use Cases

### 1. WhatsApp Service Down
Jika layanan WhatsApp sedang bermasalah dan customer tidak bisa terima OTP, admin dapat:
- Nonaktifkan OTP di settings
- Customer langsung bisa login tanpa OTP
- Aktifkan kembali OTP setelah WhatsApp normal

### 2. High Volume Login
Untuk periode dengan banyak customer login bersamaan:
- Nonaktifkan OTP untuk mengurangi beban WhatsApp API
- Hindari rate limiting
- Customer experience lebih smooth

### 3. Development/Testing
Untuk testing atau development:
- Nonaktifkan OTP untuk bypass verification
- Testing lebih cepat tanpa tunggu WhatsApp

## 🔐 Security Features

### OTP Enabled (Secure Mode)
```
Customer Input Phone → Validate User → Send OTP via WhatsApp
→ Customer Input OTP → Verify OTP → Create Session → Login Success
```

**Protection:**
- Rate limiting: Max 3 OTP per 15 menit
- OTP expiry: Default 5 menit (configurable)
- 6-digit random code
- Session verification required

### OTP Disabled (Bypass Mode)
```
Customer Input Phone → Validate User → Create Session → Login Success
```

**Protection:**
- Validate phone number exists in database
- Create verified session immediately
- Activity logging for audit trail
- IP address tracking

## 📱 Customer Login Flow

### With OTP Enabled

1. Customer buka `/login`
2. Input nomor WhatsApp atau Customer ID (8 digit)
3. System check apakah OTP enabled
4. Jika enabled: send OTP via WhatsApp
5. Customer input 6-digit OTP code
6. System verify OTP
7. Jika valid: create session, redirect ke `/customer`

### With OTP Disabled

1. Customer buka `/login`
2. Input nomor WhatsApp atau Customer ID (8 digit)
3. System check apakah OTP enabled
4. Jika disabled: langsung create session
5. Redirect ke `/customer` (no OTP required)

### Fallback: OTP Send Failed

Jika OTP enabled tapi WhatsApp gagal kirim OTP:
1. Customer lihat error message
2. Muncul button **"Masuk Tanpa OTP (WhatsApp Bermasalah)"**
3. Customer klik button bypass
4. System check apakah OTP benar-benar disabled
5. Jika OTP masih enabled: return error (security)
6. Jika OTP disabled: create session, login success

## 🛠️ Technical Implementation

### API Endpoints

#### 1. `/api/customer/auth/login` (POST)
Check user dan tentukan apakah perlu OTP atau tidak.

**Request:**
```json
{
  "identifier": "08123456789" // or "12345678" (8-digit customer ID)
}
```

**Response (OTP Disabled):**
```json
{
  "success": true,
  "otpEnabled": false,
  "requireOTP": false,
  "token": "xxx...xxx",
  "user": { ... }
}
```

**Response (OTP Enabled):**
```json
{
  "success": true,
  "otpEnabled": true,
  "requireOTP": true,
  "user": {
    "phone": "628123456789"
  },
  "token": null
}
```

#### 2. `/api/customer/auth/send-otp` (POST)
Kirim OTP code via WhatsApp (hanya jika OTP enabled).

**Request:**
```json
{
  "phone": "628123456789"
}
```

**Response:**
```json
{
  "success": true,
  "message": "OTP sent successfully",
  "expiresIn": 5
}
```

#### 3. `/api/customer/auth/verify-otp` (POST)
Verify OTP code dan create session.

**Request:**
```json
{
  "phone": "628123456789",
  "otpCode": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "token": "xxx...xxx",
  "user": { ... }
}
```

#### 4. `/api/customer/auth/bypass-login` (POST) **[NEW]**
Login tanpa OTP (hanya jika OTP disabled).

**Request:**
```json
{
  "phone": "628123456789"
}
```

**Response (OTP Disabled):**
```json
{
  "success": true,
  "token": "xxx...xxx",
  "user": { ... }
}
```

**Response (OTP Enabled - Security Block):**
```json
{
  "success": false,
  "error": "OTP is required. Please contact admin if WhatsApp service is unavailable."
}
```

### Database Schema

**whatsapp_reminder_settings**
```sql
CREATE TABLE whatsapp_reminder_settings (
  id VARCHAR PRIMARY KEY,
  enabled BOOLEAN DEFAULT TRUE,
  otpEnabled BOOLEAN DEFAULT TRUE,  -- Toggle OTP ON/OFF
  otpExpiry INT DEFAULT 5,          -- OTP expiry in minutes
  ...
);
```

**customerSession**
```sql
CREATE TABLE customer_sessions (
  id VARCHAR PRIMARY KEY,
  userId VARCHAR,
  phone VARCHAR,
  otpCode VARCHAR NULL,      -- NULL if bypass login
  otpExpiry TIMESTAMP NULL,  -- NULL if bypass login
  token VARCHAR UNIQUE,
  expiresAt TIMESTAMP,
  verified BOOLEAN DEFAULT FALSE,
  createdAt TIMESTAMP,
  updatedAt TIMESTAMP
);
```

### Frontend Logic

**src/app/login/page.tsx**
```typescript
const handleSendOTP = async () => {
  // Step 1: Check if OTP required
  const checkRes = await fetch('/api/customer/auth/login', {
    method: 'POST',
    body: JSON.stringify({ identifier })
  });
  
  const checkData = await checkRes.json();
  
  // Step 2: If OTP not required, login directly
  if (!checkData.requireOTP) {
    localStorage.setItem('customer_token', checkData.token);
    router.push('/customer');
    return;
  }
  
  // Step 3: Send OTP via WhatsApp
  const res = await fetch('/api/customer/auth/send-otp', {
    method: 'POST',
    body: JSON.stringify({ phone: userPhone })
  });
  
  const data = await res.json();
  
  // Step 4a: OTP sent successfully
  if (data.success) {
    setStep('otp');
  } 
  // Step 4b: OTP send failed - show bypass button
  else {
    setError(data.error);
    setOtpSendFailed(true); // Show bypass option
  }
};
```

## 📊 Activity Logging

Semua login activity dicatat di `activity_logs`:

**OTP Login:**
```json
{
  "module": "customer_auth",
  "action": "verify_otp",
  "description": "Customer user123@realm logged in via OTP",
  "metadata": "{\"phone\":\"628123456789\"}"
}
```

**Bypass Login:**
```json
{
  "module": "customer_auth",
  "action": "bypass_login",
  "description": "Customer user123@realm logged in without OTP (OTP disabled)",
  "metadata": "{\"phone\":\"628123456789\"}"
}
```

## ⚠️ Best Practices

### Security Recommendations

1. **Enable OTP untuk Production**
   - OTP mode lebih secure untuk production
   - Validasi nomor WhatsApp aktif

2. **Disable OTP hanya saat Darurat**
   - WhatsApp service down
   - Mass login event (promo, dll)
   - Maintenance WhatsApp API

3. **Monitor Activity Logs**
   - Track bypass login frequency
   - Detect suspicious patterns
   - Audit customer access

4. **Configure OTP Expiry**
   - Default: 5 menit
   - Recommended: 3-5 menit
   - Max: 10 menit

5. **Rate Limiting**
   - Max 3 OTP per 15 menit per phone
   - Prevent OTP spam/abuse

### Operational Guidelines

**Before Disabling OTP:**
- ✅ Verify WhatsApp service benar-benar down
- ✅ Inform customer via broadcast
- ✅ Set time limit untuk bypass mode
- ✅ Monitor login activity

**After Enabling OTP Back:**
- ✅ Test OTP delivery working
- ✅ Announce via broadcast
- ✅ Review bypass login logs

## 🐛 Troubleshooting

### Customer Cannot Login (OTP Enabled)

**Symptoms:**
- OTP tidak terkirim
- OTP expired terus
- Error "WhatsApp service unavailable"

**Solutions:**
1. Check WhatsApp API status
2. Verify WhatsApp credentials di settings
3. Check rate limiting (max 3 OTP/15min)
4. Sementara disable OTP untuk emergency access

### Customer Cannot Login (OTP Disabled)

**Symptoms:**
- Nomor valid tapi tidak bisa login
- Error "Phone not registered"

**Solutions:**
1. Verify nomor di database `pppoe_users`
2. Check phone format (62xxx vs 08xxx)
3. Verify user status = 'active'
4. Check session table untuk duplicate

### Bypass Button Not Showing

**Symptoms:**
- OTP send gagal tapi no bypass button

**Solutions:**
1. Clear browser cache
2. Check console for JavaScript errors
3. Verify API `/api/customer/auth/bypass-login` exists
4. Update frontend code

## 📞 Support

**For Admins:**
- Toggle OTP: Admin Panel > WhatsApp > Notifications
- View Logs: Admin Panel > Activity Logs > Filter "customer_auth"

**For Customers:**
- Login Page: `/login`
- Contact admin jika gagal login
- Gunakan Customer ID 8-digit sebagai alternatif

---

**Version**: 2.8.0  
**Last Updated**: December 24, 2025  
**Status**: ✅ Production Ready
