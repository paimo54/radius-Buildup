# PREPAID & POSTPAID WORKFLOW - FINAL IMPLEMENTATION

> **Updated:** January 4, 2026  
> **Version:** 4.0 - Sesuai referensi phpnuxbill  
> **Reference:** [hotspotbilling/phpnuxbill](https://github.com/hotspotbilling/phpnuxbill)

## 🎯 OVERVIEW

Sistem billing mendukung 2 jenis subscription:

- **PREPAID (Prabayar)**: Bayar dimuka, dapat masa aktif tetap
- **POSTPAID (Pascabayar)**: Tagihan bulanan, tanggal jatuh tempo tetap

**KEY INSIGHT dari phpnuxbill:**
- KEDUA jenis (PREPAID & POSTPAID) **PUNYA expiredAt**
- Perbedaannya adalah **cara menghitung expiredAt**:
  - PREPAID: expiredAt = tanggal bayar + validity
  - POSTPAID: expiredAt = billingDay bulan berikutnya

---

## 📊 POSTPAID WORKFLOW (Pascabayar)

### **Karakteristik:**
- ✅ Tagihan bulanan tetap
- ✅ **expiredAt = billingDay bulan berikutnya** (otomatis diperpanjang setelah bayar)
- ✅ **billingDay** (1-31): Tanggal jatuh tempo tagihan setiap bulan
- ✅ Invoice generate **H-7** sebelum expiredAt (sama seperti PREPAID)
- ✅ Auto-isolate jika expired DAN ada invoice OVERDUE

### **Timeline POSTPAID:**

```
┌─────────────────────────────────────────────────────────────┐
│  POSTPAID dengan billingDay = 20                            │
├─────────────────────────────────────────────────────────────┤
│  1 Jan  → User daftar POSTPAID                              │
│           billingDay: 20                                    │
│           expiredAt: 20 Feb 2026 (billingDay bulan depan)   │
│           Status: active                                    │
│                                                             │
│  13 Feb → Auto generate invoice (H-7 sebelum expiredAt)     │
│           Due Date: 20 Feb                                  │
│           Amount: Rp 200.000                                │
│           Status: PENDING                                   │
│                                                             │
│  18 Feb → User bayar invoice                                │
│           Invoice: PENDING → PAID                           │
│           expiredAt: 20 Feb → 20 Mar (perpanjang 1 bulan)   │
│                                                             │
│  13 Mar → Auto generate invoice berikutnya (H-7)            │
│           Due Date: 20 Mar                                  │
│           Amount: Rp 200.000                                │
│                                                             │
│  21 Mar → expiredAt lewat, invoice PENDING → OVERDUE        │
│                                                             │
│  22 Mar → AUTO-ISOLATE (expired + OVERDUE)                  │
│           - Status: active → isolated                       │
│           - RADIUS: masuk grup 'isolir'                     │
│           - Session: CoA disconnect                         │
│                                                             │
│  25 Mar → User bayar invoice                                │
│           - Invoice: OVERDUE → PAID                         │
│           - expiredAt: 25 Mar → 20 Apr                      │
│           - Status: isolated → active                       │
└─────────────────────────────────────────────────────────────┘
```

### **Database Schema:**
```typescript
{
  subscriptionType: 'POSTPAID',
  billingDay: 20,            // ← Tanggal jatuh tempo (1-31)
  expiredAt: DateTime,       // ← billingDay bulan berikutnya
  status: 'active',
  balance: 0,
  autoRenewal: false,        // ← N/A untuk POSTPAID
}
```

---

## 🎫 PREPAID WORKFLOW (Prabayar)

(Tetap sama seperti sebelumnya)

---

## 🔧 CONFIGURATION

### Set User sebagai POSTPAID:
```typescript
const validity = 30; // days or 1 month
const billingDay = 20; // 1-31
const expiredAt = new Date();

// Set expiredAt = billingDay bulan berikutnya
expiredAt.setMonth(expiredAt.getMonth() + 1);
expiredAt.setDate(billingDay);
expiredAt.setHours(23, 59, 59, 999);

await prisma.pppoeUser.update({
  where: { id: userId },
  data: {
    subscriptionType: 'POSTPAID',
    billingDay: billingDay,  // ← WAJIB (1-31)
    expiredAt: expiredAt,    // ← WAJIB (billingDay bulan depan)
    autoRenewal: false,      // ← N/A
  }
});
```

---

## 📅 CRON SCHEDULE

### Invoice Generation:
```
Schedule: 0 1 * * *  (Daily at 01:00 WIB)
- PREPAID: Generate untuk expiredAt H+7 sampai H+30
- POSTPAID: Generate untuk expiredAt H+7 sampai H+30 (SAMA!)
```

### Auto-Isolation:
```
Schedule: 0 * * * *  (Hourly)
- PREPAID: Isolate jika expired + has unpaid invoice
- POSTPAID: Isolate jika expired + has OVERDUE invoice
```

---

**Version:** 4.1 — Updated March 27, 2026  
**Reference:** https://github.com/hotspotbilling/phpnuxbill  
**Last Updated:** March 27, 2026  
**Status:** ✅ Sesuai implementasi v2.11.6

> ⚠️ **CATATAN PENTING (v2.11.6):** Workflow lama yang menyebutkan `expiredAt = NULL` untuk POSTPAID sudah **TIDAK BERLAKU**. Implementasi saat ini (sesuai Workflow 1 di atas) menggunakan `expiredAt = billingDay bulan berikutnya` untuk semua tipe. Saat admin mengedit `billingDay` user POSTPAID, `expiredAt` otomatis di-recalculate di `updatePppoeUser` service.

---

## 🎫 PREPAID WORKFLOW (Prabayar)

### **Karakteristik:**
- ✅ Bayar dimuka, dapat masa aktif
- ✅ **expiredAt = DateTime** (tanggal kadaluarsa)
- ✅ Invoice generate **H-7** sebelum expiredAt
- ✅ Auto-renewal: jika balance >= price, auto bayar
- ✅ Isolate: jika lewat expiredAt dan belum bayar

### **Timeline PREPAID:**

```
┌─────────────────────────────────────────────────────────────┐
│  SCENARIO 1: PREPAID NORMAL (Tanpa Auto-Renewal)           │
├─────────────────────────────────────────────────────────────┤
│  1 Jan  → User daftar, bayar paket 1 bulan                 │
│           expiredAt: 1 Feb 2026                             │
│           Status: active                                    │
│                                                             │
│  25 Jan → Auto generate invoice renewal (H-7)               │
│           Due Date: 1 Feb (expiredAt)                       │
│           Status: PENDING                                   │
│                                                             │
│  31 Jan → User bayar invoice                                │
│           expiredAt: 1 Feb → 1 Mar (diperpanjang)           │
│                                                             │
│  25 Feb → Auto generate invoice renewal berikutnya          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  SCENARIO 2: PREPAID + AUTO-RENEWAL (Balance Cukup)        │
├─────────────────────────────────────────────────────────────┤
│  1 Jan  → User daftar, top-up balance Rp 600.000           │
│           expiredAt: 1 Feb 2026                             │
│           autoRenewal: true                                 │
│           balance: 600.000                                  │
│                                                             │
│  29 Jan → Auto-renewal cron (H-3)                           │
│           - Generate invoice renewal                        │
│           - Auto bayar dari balance                         │
│           - Invoice: PENDING → PAID                         │
│           - Balance: 600.000 → 400.000                      │
│           - expiredAt: 1 Feb → 1 Mar                        │
│                                                             │
│  26 Feb → Auto-renewal cron                                 │
│           - Generate invoice renewal                        │
│           - Auto bayar dari balance                         │
│           - Balance: 400.000 → 200.000                      │
│           - expiredAt: 1 Mar → 1 Apr                        │
│                                                             │
│  26 Mar → Auto-renewal cron                                 │
│           - Generate invoice renewal                        │
│           - Balance kurang (200.000 < 200.000)              │
│           - Invoice tetap PENDING (menunggu top-up)         │
│                                                             │
│  2 Apr  → expiredAt lewat, belum bayar                      │
│           - AUTO-ISOLATE                                    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  SCENARIO 3: PREPAID Tidak Bayar (Expired)                 │
├─────────────────────────────────────────────────────────────┤
│  1 Jan  → User aktif, expiredAt: 1 Feb                      │
│                                                             │
│  25 Jan → Invoice generate (H-7)                            │
│                                                             │
│  2 Feb  → expiredAt lewat, invoice belum dibayar            │
│           - AUTO-ISOLATE                                    │
│           - Status: active → isolated                       │
│                                                             │
│  5 Feb  → User bayar invoice                                │
│           - Invoice: PENDING → PAID                         │
│           - expiredAt: 2 Feb + 30 hari = 5 Mar              │
│           - Status: isolated → active                       │
│           - RADIUS: reactivate                              │
└─────────────────────────────────────────────────────────────┘
```

### **Database Schema:**
```typescript
{
  subscriptionType: 'PREPAID',
  expiredAt: DateTime,       // ← WAJIB ada
  billingDay: null,          // ← N/A untuk PREPAID
  status: 'active',
  balance: 500000,           // Saldo untuk auto-renewal
  autoRenewal: true,         // Enable auto-renewal
}
```

### **Invoice Generation Query:**
```typescript
// Cron: Daily at 00:01 WIB
const startDate = new Date();
startDate.setDate(startDate.getDate() + 7); // H+7

const endDate = new Date();
endDate.setDate(endDate.getDate() + 30); // H+30

const prepaidUsers = await prisma.pppoeUser.findMany({
  where: {
    status: { in: ['active', 'isolated', 'blocked', 'suspended'] },
    subscriptionType: 'PREPAID',
    expiredAt: {
      gte: startDate, // 7 hari kedepan
      lte: endDate,   // sampai 30 hari kedepan
    }
  }
});

// Generate invoice dengan dueDate = user.expiredAt
```

### **Auto-Renewal Logic:**
```typescript
// Cron: Daily at 08:00 WIB
const threeDaysAhead = new Date();
threeDaysAhead.setDate(threeDaysAhead.getDate() + 3);

const usersNearExpiry = await prisma.pppoeUser.findMany({
  where: {
    subscriptionType: 'PREPAID',
    autoRenewal: true,
    expiredAt: {
      gte: new Date(),
      lte: threeDaysAhead // Expiring dalam 3 hari
    },
    balance: { gte: /* profile.price */ }
  }
});

// Auto generate invoice & auto pay dari balance
```

### **Auto-Isolation Query:**
```typescript
// Cron: Hourly
const expiredPrepaidUsers = await prisma.pppoeUser.findMany({
  where: {
    status: 'active',
    subscriptionType: 'PREPAID',
    expiredAt: { lt: new Date() } // Already expired
  },
  include: {
    invoices: {
      where: { status: { in: ['PENDING', 'OVERDUE'] } }
    }
  }
});

// Filter: Only isolate if has unpaid invoice
// (auto-renewal success = no unpaid invoice)
const toIsolate = expiredPrepaidUsers.filter(user => 
  user.invoices.length > 0 || !user.autoRenewal
);
```

---

## 🔧 CONFIGURATION

### Set User sebagai POSTPAID:
```typescript
await prisma.pppoeUser.update({
  where: { id: userId },
  data: {
    subscriptionType: 'POSTPAID',
    expiredAt: null,        // ← WAJIB null
    billingDay: 1,          // ← Tidak digunakan (always tanggal 1)
    autoRenewal: false,     // ← N/A untuk POSTPAID
  }
});
```

### Set User sebagai PREPAID:
```typescript
const validity = 30; // days
const expiredAt = new Date();
expiredAt.setDate(expiredAt.getDate() + validity);

await prisma.pppoeUser.update({
  where: { id: userId },
  data: {
    subscriptionType: 'PREPAID',
    expiredAt: expiredAt,   // ← WAJIB ada
    billingDay: null,       // ← N/A
    autoRenewal: true,      // ← Optional
    balance: 500000,        // ← Untuk auto-renewal
  }
});
```

---

## 📅 CRON SCHEDULE

### Invoice Generation:
```
Schedule: 0 1 * * *  (Daily at 01:00 WIB)
- PREPAID: Generate untuk expiredAt H+7 sampai H+30
- POSTPAID: Generate HANYA di tanggal 1 setiap bulan
```

### Invoice Status Update:
```
Schedule: 0 * * * *  (Hourly)
- Update PENDING → OVERDUE jika lewat dueDate
```

### Auto-Isolation:
```
Schedule: 0 * * * *  (Hourly)
- PREPAID: Isolate jika expired + has unpaid invoice
- POSTPAID: Isolate jika OVERDUE > 7 hari
```

### Auto-Renewal (PREPAID only):
```
Schedule: 0 8 * * *  (Daily at 08:00 WIB)
- Check users expiring dalam 3 hari
- Auto pay dari balance jika cukup
```

---

## ⚠️ VALIDASI & ERROR HANDLING

### POSTPAID Validation:
```typescript
// RULE: POSTPAID TIDAK BOLEH punya expiredAt
if (user.subscriptionType === 'POSTPAID' && user.expiredAt !== null) {
  throw new Error('POSTPAID user must not have expiredAt');
}
```

### PREPAID Validation:
```typescript
// RULE: PREPAID WAJIB punya expiredAt
if (user.subscriptionType === 'PREPAID' && !user.expiredAt) {
  throw new Error('PREPAID user must have expiredAt');
}
```

### Billing Day Validation:
```typescript
// RULE: billingDay 1-31
if (billingDay < 1 || billingDay > 31) {
  throw new Error('billingDay must be between 1-31');
}

// CATATAN: Untuk POSTPAID, billingDay diabaikan (always tanggal 1)
```

---

## 🔄 PAYMENT FLOW

### POSTPAID Payment:
```typescript
1. User bayar invoice (via payment gateway)
2. Webhook receive payment
3. Update invoice: PENDING/OVERDUE → PAID
4. Update user status: isolated → active (if isolated)
5. Sync to RADIUS: reactivate user
6. Send CoA: disconnect & reconnect
7. Send notification: WA + Email
8. ✅ Done - user dapat internet lagi
```

### PREPAID Payment:
```typescript
1. User bayar invoice (via payment gateway)
2. Webhook receive payment
3. Update invoice: PENDING/OVERDUE → PAID
4. Update expiredAt:
   - Jika belum expired: expiredAt += validity
   - Jika sudah expired: expiredAt = now + validity
5. Update user status: isolated → active (if isolated)
6. Sync to RADIUS: reactivate user
7. Send CoA: disconnect & reconnect
8. Send notification: WA + Email
9. ✅ Done - user dapat internet + masa aktif diperpanjang
```

---

## 📈 ADVANTAGES

### POSTPAID:
✅ Simple untuk user: pakai dulu bayar belakangan  
✅ Predictable: tagihan selalu tanggal 1  
✅ Grace period: 7 hari untuk bayar  
✅ Tolerance: 7 hari tambahan sebelum isolate  

### PREPAID:
✅ No debt: user bayar dimuka  
✅ Auto-renewal: jika balance cukup, auto perpanjang  
✅ Flexible: user bisa top-up kapan saja  
✅ Clear expiry: user tahu kapan masa aktif habis  

---

## 🚀 BEST PRACTICES

1. **POSTPAID**: Cocok untuk customer tetap, corporate, B2B
2. **PREPAID**: Cocok untuk customer retail, home user, yang tidak mau terikat kontrak
3. **Auto-renewal**: Aktifkan untuk user yang mau praktis (top-up sekali untuk beberapa bulan)
4. **Monitoring**: Setup alert untuk balance < 1x harga paket (untuk auto-renewal users)
5. **Communication**: Kirim reminder H-3 sebelum expiry (PREPAID) atau H-3 sebelum due date (POSTPAID)

---

**Version:** 3.0  
**Last Updated:** January 4, 2026  
**Status:** ✅ Production Ready
