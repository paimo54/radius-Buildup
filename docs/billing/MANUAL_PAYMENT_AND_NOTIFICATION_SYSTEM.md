# Manual Payment & Notification System

Sistem ini menangani alur pembayaran manual di mana customer melakukan upload bukti pembayaran dan admin melakukan review.

## Overview

Alur lengkap:
1. Customer memilih invoice yang ingin dibayar
2. Customer memilih metode pembayaran "Manual"
3. Customer upload bukti transfer
4. Admin menerima notifikasi (WhatsApp & Email)
5. Admin review bukti pembayaran
6. Admin approve atau reject
7. Customer menerima notifikasi hasil review

## Features

### 1. Customer Side

#### a. Payment Page
- **URL**: `/payment/{invoiceNumber}`
- Customer dapat melihat detail invoice
- Pilihan metode pembayaran:
  - Payment Gateway (Midtrans/Xendit/etc)
  - Manual Payment

#### b. Manual Payment Form
- Upload bukti transfer (image)
- Nama pengirim
- Bank pengirim
- Nomor rekening pengirim
- Catatan (optional)
- Preview image sebelum submit

#### c. Thank You Page
- Konfirmasi bukti sudah diupload
- Informasi bahwa pembayaran sedang direview
- Estimasi waktu review

### 2. Admin Side

#### a. Review Page
- **URL**: `/admin/payments/review`
- Table semua pembayaran dengan status MANUAL_REVIEW
- Detail:
  - Customer info
  - Invoice info
  - Bukti transfer (full size preview)
  - Bank pengirim dan nama
- Actions:
  - Approve: Update status ke PAID, extend user expiry
  - Reject: Update status ke REJECTED, tambah catatan alasan

#### b. Notification to Admin
Ketika customer upload bukti:
- **WhatsApp**: Kirim ke admin dengan detail pembayaran
- **Email**: Kirim ke admin dengan attachment bukti

### 3. Notification System

#### a. Admin Notification (New Manual Payment)
**Trigger**: Saat customer upload bukti pembayaran

**WhatsApp Template**:
```
🔔 *NOTIFIKASI PEMBAYARAN MANUAL*

📋 *Detail Pembayaran*
━━━━━━━━━━━━━━━━━━━━
📌 Invoice: {{invoiceNumber}}
👤 Pelanggan: {{customerName}}
🆔 ID: {{customerId}}
💰 Jumlah: {{amount}}

🏦 *Info Transfer*
━━━━━━━━━━━━━━━━━━━━
Bank: {{senderBank}}
Nama: {{senderName}}
No. Rek: {{senderAccount}}

📝 Catatan: {{notes}}

⏰ Waktu: {{uploadTime}}

🔗 Review: {{reviewLink}}

⚠️ Silakan cek dan verifikasi pembayaran ini.
```

**Email Template**:
- Subject: "[REVIEW NEEDED] Pembayaran Manual - {invoiceNumber}"
- Body: Detail pembayaran dengan attachment bukti

#### b. Customer Notification (Payment Approved)
**Trigger**: Saat admin approve pembayaran

**WhatsApp Template**:
```
✅ *PEMBAYARAN DIKONFIRMASI*

Halo {{customerName}},

Terima kasih! Pembayaran Anda telah kami konfirmasi.

📋 *Detail*
━━━━━━━━━━━━━━━━━━━━
📌 Invoice: {{invoiceNumber}}
💰 Jumlah: {{amount}}
✅ Status: LUNAS
📅 Aktif hingga: {{newExpiry}}

Layanan Anda telah diperpanjang.

Terima kasih telah menjadi pelanggan setia kami! 🙏

{{companyName}}
```

#### c. Customer Notification (Payment Rejected)
**Trigger**: Saat admin reject pembayaran

**WhatsApp Template**:
```
❌ *PEMBAYARAN DITOLAK*

Halo {{customerName}},

Mohon maaf, pembayaran Anda tidak dapat kami proses.

📋 *Detail*
━━━━━━━━━━━━━━━━━━━━
📌 Invoice: {{invoiceNumber}}
💰 Jumlah: {{amount}}
❌ Status: Ditolak
📝 Alasan: {{rejectionReason}}

Silakan lakukan pembayaran ulang atau hubungi kami jika ada pertanyaan.

{{companyName}}
📞 {{companyPhone}}
```

## API Endpoints

### 1. Upload Manual Payment
**Endpoint**: `POST /api/payment/manual`

**Request**:
```json
{
  "invoiceId": 123,
  "proofImage": "base64_image_data",
  "senderName": "John Doe",
  "senderBank": "BCA",
  "senderAccount": "1234567890",
  "notes": "Transfer dari rekening istri"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Bukti pembayaran berhasil diupload",
  "paymentId": 456
}
```

### 2. Get Manual Payments for Review
**Endpoint**: `GET /api/admin/payments/manual`

**Response**:
```json
{
  "payments": [
    {
      "id": 456,
      "invoiceId": 123,
      "invoiceNumber": "INV-2024-0001",
      "amount": 150000,
      "customerName": "John Doe",
      "customerId": "CUST001",
      "proofImageUrl": "/uploads/proofs/xxx.jpg",
      "senderName": "John Doe",
      "senderBank": "BCA",
      "senderAccount": "1234567890",
      "notes": "Transfer dari rekening istri",
      "uploadedAt": "2024-01-15T10:30:00Z",
      "status": "MANUAL_REVIEW"
    }
  ]
}
```

### 3. Approve Manual Payment
**Endpoint**: `POST /api/admin/payments/manual/{id}/approve`

**Request**:
```json
{
  "notes": "Pembayaran sesuai"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Pembayaran berhasil dikonfirmasi",
  "newExpiry": "2024-02-15"
}
```

### 4. Reject Manual Payment
**Endpoint**: `POST /api/admin/payments/manual/{id}/reject`

**Request**:
```json
{
  "reason": "Bukti tidak jelas, nominal tidak sesuai"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Pembayaran ditolak"
}
```

## Database Schema

### ManualPayment Table
```sql
CREATE TABLE manual_payments (
  id INT PRIMARY KEY AUTO_INCREMENT,
  invoice_id INT NOT NULL,
  proof_image_url VARCHAR(255) NOT NULL,
  sender_name VARCHAR(100) NOT NULL,
  sender_bank VARCHAR(50) NOT NULL,
  sender_account VARCHAR(50) NOT NULL,
  notes TEXT,
  status ENUM('PENDING', 'APPROVED', 'REJECTED') DEFAULT 'PENDING',
  rejection_reason TEXT,
  reviewed_by INT,
  reviewed_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (invoice_id) REFERENCES invoices(id),
  FOREIGN KEY (reviewed_by) REFERENCES users(id)
);
```

## Template Variables

### Admin Notification Template
| Variable | Description |
|----------|-------------|
| `{{invoiceNumber}}` | Nomor invoice |
| `{{customerName}}` | Nama pelanggan |
| `{{customerId}}` | ID pelanggan |
| `{{amount}}` | Jumlah pembayaran (formatted) |
| `{{senderBank}}` | Bank pengirim |
| `{{senderName}}` | Nama pengirim |
| `{{senderAccount}}` | Nomor rekening pengirim |
| `{{notes}}` | Catatan dari customer |
| `{{uploadTime}}` | Waktu upload bukti |
| `{{reviewLink}}` | Link ke halaman review |

### Customer Approved Template
| Variable | Description |
|----------|-------------|
| `{{customerName}}` | Nama pelanggan |
| `{{invoiceNumber}}` | Nomor invoice |
| `{{amount}}` | Jumlah pembayaran (formatted) |
| `{{newExpiry}}` | Tanggal expired baru |
| `{{companyName}}` | Nama perusahaan |

### Customer Rejected Template
| Variable | Description |
|----------|-------------|
| `{{customerName}}` | Nama pelanggan |
| `{{invoiceNumber}}` | Nomor invoice |
| `{{amount}}` | Jumlah pembayaran (formatted) |
| `{{rejectionReason}}` | Alasan penolakan |
| `{{companyName}}` | Nama perusahaan |
| `{{companyPhone}}` | Nomor telepon perusahaan |

## Implementation Notes

### 1. Image Upload
- Format: JPEG, PNG
- Max size: 5MB
- Storage: Local `/public/uploads/proofs/` atau S3
- Generate unique filename

### 2. Notification Priority
- WhatsApp dikirim terlebih dahulu
- Email sebagai backup/record

### 3. Security
- Validate image before save
- Admin-only access untuk review page
- Rate limiting untuk upload

### 4. Mobile Responsiveness
- Upload form mobile-friendly
- Review page responsive

## Testing Checklist

- [ ] Customer dapat upload bukti dengan berbagai format image
- [ ] Admin menerima WhatsApp notifikasi
- [ ] Admin menerima Email notifikasi
- [ ] Admin dapat approve pembayaran
- [ ] Customer menerima notifikasi approved
- [ ] User expiry terupdate setelah approved
- [ ] Admin dapat reject dengan alasan
- [ ] Customer menerima notifikasi rejected
