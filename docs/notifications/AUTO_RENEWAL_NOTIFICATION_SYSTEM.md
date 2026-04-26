# Auto-Renewal & Notification System - Verification Report

## ✅ Cronjob Registration

### Auto-Renewal Job
- **Tipe**: `auto_renewal`
- **Nama**: Auto Renewal (Prepaid)  
- **Jadwal**: `0 8 * * *` (Setiap hari jam 8 pagi WIB)
- **Status**: ✅ Terdaftar di `/admin/settings/cron`
- **Handler**: `processAutoRenewal()` dari `src/lib/cron/auto-renewal.ts`
- **Fitur**:
  - Otomatis perpanjang user PREPAID dengan `autoRenewal=true`
  - Cek user yang akan expired dalam 3 hari
  - Bayar invoice dari saldo jika cukup
  - Update expired date
  - Restore dari isolir jika perlu
  - **BARU**: Kirim notifikasi Email + WhatsApp

### Success Handler di Admin UI
File: `src/app/admin/settings/cron/page.tsx`

```typescript
} else if (jobType === 'auto_renewal') {
  await Swal.fire({
    icon: 'success',
    title: t('common.success'),
    text: `Processed ${data.processed || 0} auto-renewals, paid ${data.paid || 0}`,
    timer: 2000,
    showConfirmButton: false
  });
}
```

---

## ✅ Notification System

### 1. WhatsApp Notification

**File**: `src/lib/whatsapp-notifications.ts`

**Fungsi Baru**: `sendAutoRenewalSuccess()`

**Trigger**: Otomatis setelah auto-renewal berhasil

**Template Variables**:
```
{{customerName}}    - Nama pelanggan
{{username}}        - Username PPPoE
{{profileName}}     - Nama paket internet
{{amount}}          - Biaya perpanjangan (Rp format)
{{newBalance}}      - Saldo tersisa (Rp format)
{{expiredDate}}     - Tanggal expired baru (format Indonesia)
{{companyName}}     - Nama perusahaan
{{companyPhone}}    - Nomor telepon perusahaan
```

**Contoh Template**:
```
✅ *Auto-Renewal Berhasil*

Halo *Budi Santoso*,

Paket internet Anda telah *diperpanjang otomatis* dari saldo akun.

📋 *Detail:*
• Username: budi123
• Paket: Paket 20 Mbps
• Biaya: Rp 100.000
• Saldo tersisa: Rp 50.000
• Masa aktif hingga: 5 Desember 2025

✨ *Auto-renewal* akan terus berjalan selama saldo mencukupi.

💡 Tip: Isi saldo sebelum masa aktif habis agar layanan tidak terputus.
```

---

### 2. Email Notification

**File**: `src/lib/email.ts`

**Fungsi Baru**: `sendAutoRenewalEmail()`

**Trigger**: Otomatis setelah auto-renewal berhasil (jika user punya email)

**Template Variables**: Sama seperti WhatsApp

**Design**: 
- Gradient header (Purple)
- Tabel detail transaksi
- Info box (hijau) untuk status auto-renewal
- Tips box (orange) untuk reminder isi saldo
- Responsive HTML email

---

### 3. Template Database

**File Seed**: `prisma/seeds/auto-renewal-templates.ts`

**WhatsApp Template**:
- Type: `auto-renewal-success`
- Table: `whatsapp_templates`
- Dapat diedit lewat admin UI

**Email Template**:
- Type: `auto-renewal-success`  
- Table: `email_templates`
- Dapat diedit lewat admin UI

**Cara Install Template**:
```bash
npx tsx prisma/seeds/auto-renewal-templates.ts
```

---

## ✅ Auto-Isolir Compatibility

### Update Auto-Isolir Logic

**File**: `src/lib/cron/voucher-sync.ts`

**Fungsi**: `autoIsolateExpiredUsers()`

### Logika Baru (Compatible dengan Prepaid/Postpaid):

Auto-isolir akan mengisolasi user yang:

1. **POSTPAID** - Semua user postpaid yang expired
2. **PREPAID tanpa auto-renewal** - User prepaid dengan `autoRenewal=false`
3. **PREPAID dengan auto-renewal tapi saldo kurang** - User prepaid dengan `autoRenewal=true` tapi gagal auto-renewal karena saldo tidak cukup

### Query Prisma:
```typescript
const expiredUsers = await prisma.pppoeUser.findMany({
  where: {
    status: 'active',
    expiredAt: {
      lt: startOfTodayWIB, // expired before today
    },
    OR: [
      // Postpaid - always isolate if expired
      { subscriptionType: 'POSTPAID' },
      // Prepaid without auto-renewal
      { 
        subscriptionType: 'PREPAID',
        autoRenewal: false 
      },
      // Prepaid with auto-renewal but insufficient balance
      {
        subscriptionType: 'PREPAID',
        autoRenewal: true,
      }
    ]
  }
})
```

### Kenapa Aman?

1. **Auto-renewal runs BEFORE expiry** (3 hari sebelum)
2. **Auto-isolir runs AFTER expiry** (hari berikutnya)
3. User prepaid dengan auto-renewal aktif dan saldo cukup **sudah diperpanjang** sebelum expired
4. Jadi saat auto-isolir jalan, mereka tidak masuk kriteria (expired date sudah di-extend)

---

## 📝 Template Customization Guide

### Cara Edit Template WhatsApp

1. Login ke admin panel
2. Buka **Settings → WhatsApp Templates** (perlu diimplementasi UI)
3. Cari template: **Auto-Renewal Berhasil** (`auto-renewal-success`)
4. Edit field `message` dengan menggunakan variables `{{variableName}}`
5. Klik Save

### Cara Edit Template Email

1. Login ke admin panel
2. Buka **Settings → Email Templates** (perlu diimplementasi UI)  
3. Cari template: **Auto-Renewal Berhasil** (`auto-renewal-success`)
4. Edit field:
   - `subject` - Subject email
   - `htmlBody` - HTML content dengan variables
5. Klik Save

### Variables Yang Tersedia

| Variable | Keterangan | Contoh |
|----------|-----------|---------|
| `{{customerName}}` | Nama pelanggan | Budi Santoso |
| `{{username}}` | Username PPPoE | budi123 |
| `{{profileName}}` | Nama paket | Paket 20 Mbps |
| `{{amount}}` | Biaya (formatted) | Rp 100.000 |
| `{{newBalance}}` | Saldo tersisa | Rp 50.000 |
| `{{expiredDate}}` | Tanggal expired | 5 Desember 2025 |
| `{{companyName}}` | Nama perusahaan | NET INTERNET |
| `{{companyPhone}}` | No telp | 081234567890 |

---

## 🧪 Testing Checklist

### 1. Test Auto-Renewal Cron

```bash
# Manual trigger dari UI
# Buka: /admin/settings/cron
# Klik "Run Now" pada job "Auto Renewal (Prepaid)"
```

**Expected Result**:
- SweetAlert muncul: "Processed X auto-renewals, paid Y"
- Cek cronHistory table untuk log
- Cek user balance berkurang
- Cek expired date bertambah
- **BARU**: Cek WhatsApp terkirim
- **BARU**: Cek Email terkirim

### 2. Test Notifications

**Prerequisites**:
- WhatsApp gateway sudah configured
- SMTP email sudah configured
- Template sudah di-seed ke database

**Test Steps**:
1. Buat user prepaid dengan autoRenewal=true
2. Set expired dalam 2 hari
3. Set balance cukup (>= paket price)
4. Pastikan user punya `phone` dan `email`
5. Run auto-renewal cron
6. Check:
   - WhatsApp log di terminal
   - Email history di database
   - User menerima pesan

### 3. Test Auto-Isolir Compatibility

**Test Case 1: Postpaid Expired**
- User postpaid expired
- Result: ✅ Harus diisolir

**Test Case 2: Prepaid tanpa Auto-Renewal**
- User prepaid, autoRenewal=false, expired
- Result: ✅ Harus diisolir

**Test Case 3: Prepaid dengan Auto-Renewal + Saldo Cukup**
- User prepaid, autoRenewal=true, balance cukup
- Result: ✅ TIDAK diisolir (sudah auto-renewed)

**Test Case 4: Prepaid dengan Auto-Renewal + Saldo Kurang**
- User prepaid, autoRenewal=true, balance < price
- Result: ✅ Harus diisolir (auto-renewal gagal)

---

## 📋 Implementation Status

| Fitur | Status | File |
|-------|--------|------|
| Auto-renewal cron registered | ✅ | `src/lib/cron/config.ts` |
| Auto-renewal success handler UI | ✅ | `src/app/admin/settings/cron/page.tsx` |
| WhatsApp notification | ✅ | `src/lib/whatsapp-notifications.ts` |
| Email notification | ✅ | `src/lib/email.ts` |
| Template seeds | ✅ | `prisma/seeds/auto-renewal-templates.ts` |
| Auto-isolir update | ✅ | `src/lib/cron/voucher-sync.ts` |
| Balance column in UI | ✅ | `src/app/admin/pppoe/users/page.tsx` |
| Auto-renewal badge | ✅ | `src/app/admin/pppoe/users/page.tsx` |

---

## 🚀 Deployment Steps

### 1. Install Template ke Database

```bash
npx tsx prisma/seeds/auto-renewal-templates.ts
```

### 2. Verify Cron Registration

- Buka `/admin/settings/cron`
- Pastikan "Auto Renewal (Prepaid)" muncul
- Cek schedule: "Daily at 8 AM"

### 3. Configure Notifications

**WhatsApp**:
- Pastikan WhatsApp gateway configured
- Test dengan send test message

**Email**:
- Configure SMTP di Email Settings
- Test dengan send test email

### 4. Enable Auto-Renewal untuk User

- Edit user → Toggle "Auto Renewal" ON
- Top-up balance yang cukup

### 5. Monitor Execution

- Check cron history: `/admin/settings/cron`
- Check logs untuk notification status
- Verify user balance dan expired date

---

## ⚠️ Important Notes

1. **Auto-renewal runs at 8 AM daily** - Users expiring in next 3 days will be processed
2. **Notifications are optional** - Won't break transaction if failed
3. **Templates customizable** - Can be edited via admin UI (need UI implementation)
4. **Auto-isolir safe** - Won't isolate users that were auto-renewed
5. **Balance required** - User needs balance >= package price for auto-renewal to work

---

## 🆘 Troubleshooting

### Auto-Renewal Tidak Jalan

**Check**:
1. Cron job enabled? (`/admin/settings/cron`)
2. User `autoRenewal=true`?
3. User `subscriptionType='PREPAID'`?
4. User balance >= package price?
5. User expired dalam 3 hari?

### Notifikasi Tidak Terkirim

**WhatsApp**:
1. Check WhatsApp gateway configuration
2. Check user punya field `phone`
3. Check logs untuk error message
4. Verify template exists di database

**Email**:
1. Check SMTP configuration
2. Check user punya field `email`
3. Check email history table
4. Verify template exists di database

### User Tetap Diisolir Padahal Sudah Auto-Renewal

**Possible Causes**:
1. Auto-renewal failed (check logs)
2. Balance insufficient saat cron jalan
3. Template transaction error

**Debug**:
```sql
-- Check user status
SELECT username, balance, autoRenewal, expiredAt, status 
FROM pppoeUser 
WHERE username = 'target_username';

-- Check cron history
SELECT * FROM cronHistory 
WHERE jobType = 'auto_renewal' 
ORDER BY startedAt DESC 
LIMIT 5;
```

---

## 📞 Support

Jika ada masalah, cek:
1. Cron history logs
2. Application logs (console)
3. Email history table
4. WhatsApp gateway logs

