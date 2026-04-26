# Broadcast Notification System

Sistem untuk mengirim notifikasi massal ke multiple user melalui WhatsApp dan Email.

> **Last Updated:** April 5, 2026 — Kirimi.id native broadcast API, webhook pesan masuk, per-provider error detail.

## Overview

Sistem ini memungkinkan admin untuk mengirim notifikasi ke banyak user sekaligus melalui dropdown menu dengan 3 pilihan:
1. **Notifikasi Gangguan** - Informasi maintenance/gangguan jaringan
2. **Kirim Invoice** - Reminder invoice yang belum dibayar
3. **Bukti Pembayaran** - Konfirmasi pembayaran sudah diterima

## Broadcast via WhatsApp Send Page

Selain notifikasi terstruktur di atas, admin dapat mengirim pesan bebas via **Admin → WhatsApp → Send Message** (tab Broadcast):

1. Pilih user dari tabel (multi-select, select all, filter ODP)
2. Ketik pesan (atau pilih template)
3. Klik "Broadcast" → konfirmasi jumlah user
4. Lihat hasil: `✅ Berhasil: N | ❌ Gagal: M`

### Cara Kerja per Provider

| Provider | Mekanisme Broadcast |
|----------|----------------------|
| **Kirimi.id** | Native `/v1/broadcast-message` — 1 API call per grup pesan identik. 1 penerima otomatis pakai `/v1/send-message`. Delay **30 detik** antar pesan. |
| Fonnte | Loop satu-per-satu dengan rate limiter |
| Wablas | Loop satu-per-satu dengan rate limiter |
| WAHA / MPWA / GOWA / WABlast | Loop satu-per-satu dengan rate limiter |

### Catatan Status Kirimi.id

Status **"menunggu"** di dashboard Kirimi.id adalah **normal** — artinya pesan sudah diterima sistem Kirimi.id dan sedang dalam antrian untuk dikirim satu per satu dengan delay 30 detik. Estimasi:

| Jumlah Penerima | Estimasi Selesai |
|-----------------|-----------------|
| 10 user | ~5 menit |
| 50 user | ~25 menit |
| 100 user | ~50 menit |

## Features

### 1. Dropdown Menu di PPPoE Users Page
- Button "Kirim Notifikasi" dengan dropdown
- 3 pilihan: Gangguan, Invoice, Bukti Pembayaran
- Dynamic form sesuai jenis notifikasi

### 2. Multi-Channel Notification
- WhatsApp
- Email
- Both (WhatsApp & Email)

### 3. Bulk Selection
- Pilih multiple users dari table
- Select all option
- Filter by status (Active, Inactive, Expired)

## Usage Flow

### A. Notifikasi Gangguan (Outage)

**Use Case:** Memberitahu customer tentang maintenance atau gangguan jaringan

**Steps:**
1. Buka halaman Admin → PPPoE Users
2. Pilih user yang ingin dikirim notifikasi (checkbox)
3. Klik "Kirim Notifikasi" → Pilih "Notifikasi Gangguan"
4. Isi form:
   - **Jenis Gangguan**: Dropdown (Gangguan Jaringan, Maintenance, Upgrade, Lainnya)
   - **Deskripsi**: Text area (jelaskan detail gangguan)
   - **Estimasi Waktu**: Input waktu (contoh: "2-3 jam", "1 hari")
   - **Area Terdampak**: Input text (contoh: "Seluruh area", "Area A")
5. Pilih metode notifikasi: WhatsApp, Email, atau Both
6. Klik "Kirim"

**Template Variables:**
- `{{customerName}}` - Nama customer
- `{{customerId}}` - ID customer
- `{{username}}` - Username PPPoE
- `{{issueType}}` - Jenis gangguan
- `{{description}}` - Deskripsi gangguan
- `{{estimatedTime}}` - Estimasi waktu penyelesaian
- `{{affectedArea}}` - Area terdampak
- `{{companyName}}` - Nama perusahaan
- `{{companyPhone}}` - Telepon perusahaan
- `{{companyEmail}}` - Email perusahaan

### B. Kirim Invoice

**Use Case:** Mengirim reminder invoice yang belum dibayar ke customer

**Steps:**
1. Buka halaman Admin → PPPoE Users
2. Pilih user yang ingin dikirim invoice (checkbox)
3. Klik "Kirim Notifikasi" → Pilih "Kirim Invoice"
4. (Optional) Tambahkan pesan tambahan
5. Pilih metode notifikasi: WhatsApp, Email, atau Both
6. Klik "Kirim"

**Logic:**
- Sistem akan fetch invoice terakhir untuk setiap user
- Hanya user yang memiliki invoice yang akan dikirimi notifikasi
- User tanpa invoice akan skip dengan error message

**Template Variables:**
- `{{customerName}}` - Nama customer
- `{{customerId}}` - ID customer
- `{{username}}` - Username PPPoE
- `{{invoiceNumber}}` - Nomor invoice
- `{{amount}}` - Total tagihan (formatted Rupiah)
- `{{dueDate}}` - Tanggal jatuh tempo
- `{{profileName}}` - Nama paket
- `{{customerEmail}}` - Email customer
- `{{paymentLink}}` - Link pembayaran
- `{{additionalMessage}}` - Pesan tambahan (optional)
- `{{companyName}}` - Nama perusahaan
- `{{companyPhone}}` - Telepon perusahaan
- `{{companyEmail}}` - Email perusahaan
- `{{baseUrl}}` - Base URL aplikasi

### C. Bukti Pembayaran

**Use Case:** Mengirim konfirmasi pembayaran sudah diterima ke customer

**Steps:**
1. Buka halaman Admin → PPPoE Users
2. Pilih user yang sudah melakukan pembayaran (checkbox)
3. Klik "Kirim Notifikasi" → Pilih "Bukti Pembayaran"
4. (Optional) Tambahkan pesan tambahan
5. Pilih metode notifikasi: WhatsApp, Email, atau Both
6. Klik "Kirim"

**Logic:**
- Sistem akan fetch invoice terakhir yang sudah PAID untuk setiap user
- Hanya user dengan invoice PAID yang akan dikirimi notifikasi
- User tanpa invoice PAID akan skip dengan error message

**Template Variables:**
- `{{customerName}}` - Nama customer
- `{{customerId}}` - ID customer
- `{{username}}` - Username PPPoE
- `{{invoiceNumber}}` - Nomor invoice
- `{{amount}}` - Total pembayaran (formatted Rupiah)
- `{{paidDate}}` - Tanggal pembayaran
- `{{profileName}}` - Nama paket
- `{{customerEmail}}` - Email customer
- `{{expiredDate}}` - Tanggal expired akun (setelah extend)
- `{{additionalMessage}}` - Pesan tambahan (optional)
- `{{companyName}}` - Nama perusahaan
- `{{companyPhone}}` - Telepon perusahaan
- `{{companyEmail}}` - Email perusahaan
- `{{baseUrl}}` - Base URL aplikasi

## API Endpoint

**Endpoint:** `POST /api/pppoe/users/send-notification`

**Request Body:**
```json
{
  "userIds": [1, 2, 3],
  "notificationType": "outage|invoice|payment",
  "notificationMethod": "whatsapp|email|both",
  
  // For notificationType = 'outage'
  "issueType": "Gangguan Jaringan",
  "description": "Detail gangguan...",
  "estimatedTime": "2-3 jam",
  "affectedArea": "Seluruh area"
}
```

**Response Success:**
```json
{
  "message": "Notifikasi berhasil dikirim",
  "totalSent": 10,
  "emailSent": 10,
  "whatsappSent": 10,
  "failedCount": 0
}
```

**Response Error:**
```json
{
  "error": "Error message",
  "failedCount": 2,
  "errors": [
    "User John tidak memiliki invoice",
    "User Jane belum melakukan pembayaran"
  ]
}
```

## Technical Implementation

### Frontend (page.tsx)

**State Management:**
```typescript
const [notificationType, setNotificationType] = useState<'outage' | 'invoice' | 'payment'>('outage');
const [showNotificationMenu, setShowNotificationMenu] = useState(false);
const [isBroadcastDialogOpen, setIsBroadcastDialogOpen] = useState(false);
```

**Dropdown Menu:**
```typescript
const handleOpenNotificationMenu = (type: 'outage' | 'invoice' | 'payment') => {
  setNotificationType(type);
  setShowNotificationMenu(false);
  setIsBroadcastDialogOpen(true);
};
```

**Dynamic Form Rendering:**
- Outage: Form dengan 4 field (issueType, description, estimatedTime, affectedArea)
- Invoice: Info box + optional message field
- Payment: Info box + optional message field

### Backend (route.ts)

**Flow:**
1. Validate request body
2. Check notificationType
3. Fetch users with their invoices
4. Determine template type based on notificationType
5. Get email and WhatsApp templates
6. Loop through users
7. Replace template variables
8. Send via EmailService and/or WhatsApp API
9. Return success/error summary

## Best Practices

1. **Select Users Carefully:** Pastikan user yang dipilih sesuai dengan jenis notifikasi
2. **Test First:** Test dengan 1-2 user dulu sebelum send ke banyak user
3. **Check Templates:** Pastikan template sudah sesuai dan variabel sudah benar
4. **Monitor Errors:** Cek error messages jika ada user yang gagal
5. **Use Appropriate Channel:** 
   - WhatsApp: Lebih cepat, real-time
   - Email: Lebih formal, ada record
   - Both: Maksimal reach

## Troubleshooting

### Problem: User tidak menerima notifikasi
**Solution:**
- Cek phone number dan email user sudah benar
- Cek WhatsApp provider status
- Cek email settings
- Cek template exists

### Problem: Error "User tidak memiliki invoice" (Invoice type)
**Solution:**
- User belum punya invoice, generate invoice dulu
- Atau jangan include user tersebut di selection

### Problem: Error "User belum melakukan pembayaran" (Payment type)
**Solution:**
- User belum bayar invoice
- Atau mark invoice as paid dulu secara manual
