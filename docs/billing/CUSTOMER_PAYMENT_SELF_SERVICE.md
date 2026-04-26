# Customer Self-Service Payment & Package Upgrade

## Overview
Sistem self-service untuk customer yang memungkinkan pembayaran tagihan dan upgrade paket langsung dari dashboard customer tanpa perlu menghubungi admin.

## Fitur yang Ditambahkan

### 1. **Payment Gateway Integration**
- Generate payment link otomatis untuk invoice yang belum dibayar
- Support untuk multiple payment gateway:
  - **Midtrans** (Sandbox & Production)
  - **Xendit** (Invoice API)
- Auto-redirect ke payment page
- Payment link disimpan di database untuk digunakan ulang

### 2. **Invoice Payment Button**
Customer dashboard menampilkan tombol pembayaran:
- **"Bayar Sekarang"** - Jika payment link sudah ada
- **"Generate Link Pembayaran"** - Jika payment link belum ada
- Button disabled saat generating payment link (loading state)
- Auto-open payment link di tab baru

### 3. **Package Upgrade Self-Service**
- Tombol **"Ganti Paket"** di customer dashboard
- Redirect ke `/customer/wifi?tab=upgrade` untuk memilih paket baru
- Flow lengkap:
  1. Pilih paket baru
  2. System create invoice otomatis
  3. Generate payment link
  4. Customer bayar via payment gateway
  5. Auto-upgrade setelah pembayaran confirmed

### 4. **PPPoE Connection Status**
Dashboard customer sekarang menampilkan:
- **Status ONT**: Online/Offline
- **Status PPPoE**: Connected/Disconnected
- **IP PPPoE**: IP address dari PPPoE session
- **RX Power**: Signal strength
- **Temperature**: ONT temperature
- **Uptime**: Device uptime
- **Connected Devices**: Jumlah device terhubung

## API Endpoints

### `/api/customer/invoices/payment`
**Method**: `GET`

**Query Parameters**:
- `invoiceId` - Invoice ID yang akan dibayar

**Response**:
```json
{
  "success": true,
  "paymentLink": "https://app.midtrans.com/snap/v2/...",
  "invoice": {
    "id": "inv_xxx",
    "invoiceNumber": "INV-2025-001",
    "amount": 300000,
    "dueDate": "2025-01-15"
  }
}
```

**Error Responses**:
- `401 Unauthorized` - Token tidak valid
- `400 Bad Request` - Invoice ID tidak diberikan
- `404 Not Found` - Invoice tidak ditemukan atau sudah dibayar
- `503 Service Unavailable` - Payment gateway tidak dikonfigurasi
- `500 Internal Server Error` - Gagal generate payment link

## Payment Gateway Configuration

### Midtrans Setup
1. Login ke Midtrans Dashboard
2. Dapatkan **Server Key** dari Settings → Access Keys
3. Masukkan ke Admin Panel → Settings → Payment Gateway
4. Pilih environment (Sandbox/Production)

### Xendit Setup
1. Login ke Xendit Dashboard
2. Dapatkan **API Key** dari Settings → Developers → API Keys
3. Masukkan ke Admin Panel → Settings → Payment Gateway
4. Set callback URL untuk payment notification

## Database Schema Updates

### Invoice Table
```prisma
model Invoice {
  id            String   @id @default(uuid())
  invoiceNumber String   @unique
  userId        String
  amount        Float
  status        String   // PAID, UNPAID
  dueDate       DateTime
  paidAt        DateTime?
  paymentLink   String?  // NEW: Store payment link
  description   String?
  createdAt     DateTime @default(now())
  updatedAt     DateTime @updatedAt
  
  user          PppoeUser @relation(fields: [userId], references: [id])
}
```

### PaymentGatewaySetting Table
```prisma
model PaymentGatewaySetting {
  id            String   @id @default(uuid())
  gatewayName   String   // MIDTRANS, XENDIT
  apiKey        String
  serverKey     String?  // For Midtrans
  isSandbox     Boolean  @default(false)
  isActive      Boolean  @default(true)
  createdAt     DateTime @default(now())
  updatedAt     DateTime @updatedAt
}
```

## Customer Dashboard Features

### 1. Informasi Akun
- Nama, No HP, Username
- Paket aktif dengan bandwidth
- Status akun (Aktif/Expired)
- Tanggal expired dengan countdown hari

### 2. ONT/WiFi Information
- Model ONT (Manufacturer + Model)
- Status ONT (Online/Offline)
- **PPPoE Connection Status** (Connected/Disconnected)
- IP Address PPPoE
- RX Power (Signal strength)
- Temperature ONT
- Device uptime
- Jumlah device terhubung
- WiFi SSID dan status

### 3. Tagihan (Invoices)
- List semua invoice (Lunas/Belum Bayar/Terlambat)
- Invoice number dan due date
- Amount dalam Rupiah
- Status badge dengan warna:
  - **Hijau**: Lunas
  - **Kuning**: Belum Bayar
  - **Merah**: Terlambat
- Tombol pembayaran:
  - **Bayar Sekarang**: Jika payment link ada
  - **Generate Link Pembayaran**: Jika belum ada payment link

### 4. Package Upgrade Button
- Tombol "Ganti Paket" di dashboard
- Redirect ke halaman WiFi management tab upgrade
- Proses upgrade lengkap dengan payment integration

## User Flow

### Payment Flow
1. Customer login ke dashboard
2. Lihat invoice yang belum dibayar
3. Klik "Generate Link Pembayaran" (jika belum ada link)
4. System call API `/api/customer/invoices/payment`
5. API create transaction di payment gateway
6. Payment link disimpan ke database
7. Link otomatis terbuka di tab baru
8. Customer bayar via payment gateway
9. Payment gateway kirim callback ke system
10. System update invoice status menjadi PAID
11. Auto-extend expiry date customer

### Package Upgrade Flow
1. Customer klik tombol "Ganti Paket"
2. Redirect ke halaman WiFi management
3. Tab "Upgrade Paket" menampilkan list paket available
4. Customer pilih paket baru
5. System create invoice baru dengan amount = harga paket
6. Auto-generate payment link
7. Redirect ke payment page
8. Setelah bayar, package otomatis ter-upgrade

## Error Handling

### Payment Generation Errors
- **No Payment Gateway**: Tampilkan pesan "Payment gateway belum dikonfigurasi"
- **Invalid Invoice**: Tampilkan "Invoice tidak ditemukan atau sudah dibayar"
- **Gateway Error**: Tampilkan "Gagal generate payment link, coba lagi"
- **Network Error**: Tampilkan "Koneksi gagal, periksa internet Anda"

### PPPoE Status Check
- Jika `pppUsername` kosong → Status: Disconnected
- Jika `pppUsername` ada → Status: Connected
- Menggunakan VirtualParameters dari GenieACS

## Security Features

### Token Verification
- Semua API menggunakan CustomerSession authentication
- Token expiry check setiap request
- Auto-logout jika token expired

### Invoice Ownership
- Verify invoice belongs to logged-in customer
- Prevent payment for other customer's invoice
- Check invoice status before generating payment link

### Payment Gateway Security
- Use server-side API calls only
- Never expose API keys to client
- Validate payment callback signature
- Store payment link securely in database

## Testing Checklist

### Dashboard Display
- [ ] PPPoE status shows correctly (Connected/Disconnected)
- [ ] ONT information displays with VirtualParameters data
- [ ] Invoice list shows all invoices with correct status
- [ ] "Ganti Paket" button redirects to upgrade page
- [ ] Payment buttons appear only for unpaid invoices

### Payment Link Generation
- [ ] Generate button shows loading state
- [ ] Payment link opens in new tab
- [ ] Invoice updates with payment link after generation
- [ ] "Bayar Sekarang" button appears after link generated
- [ ] Error messages shown for failed generation

### Package Upgrade
- [ ] Redirect to WiFi page with upgrade tab
- [ ] Available packages displayed correctly
- [ ] Invoice created after package selection
- [ ] Payment link generated automatically
- [ ] Package upgraded after payment confirmed

## Troubleshooting

### Payment Link Not Generated
1. Check payment gateway settings in admin panel
2. Verify API keys and server keys are correct
3. Check payment gateway sandbox/production mode
4. Review server logs for API errors
5. Test payment gateway API manually

### PPPoE Status Not Showing
1. Verify GenieACS connection
2. Check if device has VirtualParameters configured
3. Verify pppUsername virtual parameter exists
4. Check device last inform time
5. Review ONT API response structure

### Package Upgrade Failed
1. Check if invoice was created successfully
2. Verify payment link generation
3. Check payment callback configuration
4. Review package upgrade API logs
5. Verify customer profile update logic

## Next Steps

1. **Auto-Payment Callback**
   - Implement webhook endpoint for payment notifications
   - Auto-update invoice status when payment confirmed
   - Send WhatsApp notification after successful payment

2. **Payment History**
   - Show payment transaction history
   - Display payment receipt/proof
   - Download invoice PDF

3. **Promo Code System**
   - Apply discount/promo code at checkout
   - Validate promo code availability
   - Calculate discount amount

4. **Recurring Payment**
   - Setup auto-debit for monthly payment
   - Save payment method (card/e-wallet)
   - Auto-charge before expiry date

## References

- Midtrans Documentation: https://docs.midtrans.com
- Xendit Invoice API: https://developers.xendit.co/api-reference/#create-invoice
- GenieACS VirtualParameters: https://github.com/genieacs/genieacs/wiki/Virtual-Parameters
- Prisma Client: https://www.prisma.io/docs/concepts/components/prisma-client
