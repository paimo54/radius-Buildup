# Solusi CoA Disconnect untuk Hotspot Voucher Expired

## Analisis Masalah

### Status Saat Ini
- âœ… Network connectivity: VPS dapat ping dan reach port 3799 UDP Mikrotik
- âœ… CoA enabled di Mikrotik: `/radius incoming accept=yes port=3799`
- âœ… Secret benar: `your-nas-secret`
- âŒ **Mikrotik tidak merespon CoA request:** `No reply from server`

### Root Cause Analysis

Masalah **"No reply from server"** biasanya disebabkan oleh:

1. **Secret di RADIUS Incoming berbeda dengan RADIUS Client**
   - Mikrotik punya 2 tempat konfigurasi RADIUS dengan secret berbeda:
     - `/radius` â†’ secret untuk **outgoing** (Mikrotik â†’ FreeRADIUS)
     - `/radius incoming` â†’ secret untuk **incoming** (FreeRADIUS â†’ Mikrotik)
   - **Keduanya HARUS pakai secret yang SAMA!**

2. **RADIUS Incoming belum fully initialized**
   - Perlu restart Mikrotik atau service RADIUS
   - Beberapa versi RouterOS butuh reboot setelah enable incoming

3. **Attribute mismatch**
   - Mikrotik strict pada attribute matching
   - Butuh Calling-Station-Id (MAC address) untuk beberapa kasus

---

## âœ… SOLUSI EFEKTIF (Tanpa API)

### LANGKAH 1: Verifikasi Secret di Mikrotik

**PENTING:** Secret di `/radius` dan `/radius incoming` HARUS SAMA!

```routeros
# 1. Cek secret di RADIUS Client (outgoing)
/radius print detail

# Output contoh:
# 0   service=hotspot,pppoe address=103.151.140.110 
#     secret="your-nas-secret" timeout=5s00ms

# 2. Set secret di RADIUS Incoming (incoming) - HARUS SAMA!
/radius incoming
set accept=yes
set default-secret="your-nas-secret"
print

# Output yang BENAR:
#   accept: yes
#     port: 3799
# default-secret: your-nas-secret
```

**Catatan Penting:**
- Jika `/radius incoming` tidak punya parameter `default-secret`, tambahkan manual
- Secret harus **PERSIS SAMA** dengan secret di `/radius`
- Setelah set secret, **RESTART Mikrotik** agar efektif

### LANGKAH 2: Restart Mikrotik (PENTING!)

```routeros
# Restart RADIUS service atau reboot Mikrotik
/system reboot
```

Banyak kasus CoA tidak bekerja karena **RADIUS Incoming belum fully initialized** setelah konfigurasi. Restart Mikrotik akan memastikan semua service berjalan dengan benar.

### LANGKAH 3: Test Ulang CoA Setelah Restart

Di VPS, jalankan:

```bash
ssh root@103.151.140.110 /root/test-coa.sh
```

**Jika SUKSES, output:**
```
Received Disconnect-ACK Id 236 from 10.20.30.11:3799
âœ“ CoA command sent successfully
```

**Jika MASIH GAGAL,** lanjut ke Step 4.

---

### LANGKAH 4: Tambahkan Calling-Station-Id (MAC Address)

Beberapa versi Mikrotik membutuhkan **MAC address** sebagai identifier untuk CoA.

Update file `src/lib/services/coaService.ts`:

```typescript
// Tambahkan Calling-Station-Id jika ada MAC address
const coaAttributes = [
  `NAS-IP-Address=${nasIpAddress}`,
  `Framed-IP-Address=${framedIpAddress}`,
  `User-Name=${username}`,
]

if (acctSessionId) {
  coaAttributes.push(`Acct-Session-Id=${acctSessionId}`)
}

// TAMBAHAN: Coba cari MAC address dari radacct
if (acctSessionId) {
  const session = await prisma.radacct.findFirst({
    where: { acctsessionid: acctSessionId },
    select: { callingstationid: true }
  })
  
  if (session?.callingstationid) {
    coaAttributes.push(`Calling-Station-Id=${session.callingstationid}`)
  }
}
```

---

### LANGKAH 5: Gunakan DM (Disconnect-Message) Code

Beberapa Mikrotik lebih responsif terhadap **code 40 (Disconnect-Request)** dengan packet type yang spesifik.

Update command di `coaService.ts`:

```typescript
// Gunakan code 40 explicitly
const command = `radclient -t 3 -r 3 -x ${nasIpAddress}:3799 40 ${nasSecret} < ${tmpFile} 2>&1`
```

Parameter:
- `-t 3` = timeout 3 detik
- `-r 3` = retry 3 kali
- `40` = Disconnect-Request code (explicit)

---

## ðŸ”§ SOLUSI ALTERNATIF (Jika CoA Tetap Gagal)

### Opsi A: Hybrid Session Timeout + CoA

Jika Mikrotik tetap tidak merespon CoA, gunakan **Session Timeout** di Hotspot Profile sebagai backup:

**Di Mikrotik:**
```routeros
# Set session timeout di RADIUS reply
# File: /etc/freeradius/mods-available/sql
```

**Update radgroupreply saat sync voucher:**
```typescript
// Di src/lib/hotspot-radius-sync.ts
await prisma.radgroupreply.create({
  data: {
    groupname: uniqueGroupName,
    attribute: 'Session-Timeout',
    op: ':=',
    value: String(voucher.profile.validityMinutes * 60) // dalam detik
  }
})
```

**Keuntungan:**
- Session otomatis disconnect setelah timeout (tanpa CoA)
- FreeRADIUS yang kontrol, bukan aplikasi
- Lebih reliable untuk network yang unstable

**Kekurangan:**
- User bisa tetap online beberapa menit setelah expired
- Tidak instant seperti CoA

---

### Opsi B: Periodic Session Validation

Tambahkan cron job yang **mark session sebagai stop** di database jika voucher expired:

```typescript
// Di src/lib/cron/voucher-sync.ts

// Setelah mark voucher as EXPIRED
if (activeSession) {
  // Update radacct langsung (tanpa CoA)
  await prisma.radacct.update({
    where: { radacctid: activeSession.radacctid },
    data: {
      acctstoptime: new Date(),
      acctterminatecause: 'Admin-Reset',
      acctsessiontime: Math.floor((Date.now() - new Date(activeSession.acctstarttime).getTime()) / 1000)
    }
  })
  
  console.log(`[CRON] âœ“ Marked session ${voucher.code} as stopped in database`)
}
```

**Update UI untuk check acctterminatecause:**

Di query active sessions, tambahkan filter:
```typescript
where: {
  acctstoptime: null,
  acctterminatecause: { not: 'Admin-Reset' } // Exclude manual stop
}
```

**Keuntungan:**
- Session hilang dari UI immediately
- Tidak perlu CoA bekerja
- Database tetap akurat

**Kekurangan:**
- User fisik masih bisa online (Mikrotik belum disconnect)
- Butuh Mikrotik Idle Timeout untuk force disconnect
- Accounting tidak 100% akurat

---

### Opsi C: Force MAC Binding + Session Count Limit

Jika CoA tidak critical, gunakan **pencegahan** bukan disconnect:

1. **Lock MAC Address:**
   ```typescript
   // Di radcheck, tambahkan MAC binding
   await prisma.radcheck.create({
     data: {
       username: voucher.code,
       attribute: 'Calling-Station-Id',
       op: ':=',
       value: macAddress // dari first login
     }
   })
   ```

2. **Limit Concurrent Sessions:**
   ```typescript
   await prisma.radcheck.create({
     data: {
       username: voucher.code,
       attribute: 'Simultaneous-Use',
       op: ':=',
       value: '1'
     }
   })
   ```

**Keuntungan:**
- User tidak bisa login lagi setelah expired (radcheck deleted)
- Mencegah abuse (1 voucher = 1 device)
- Tidak butuh CoA

**Kekurangan:**
- Session aktif tetap jalan sampai user logout manual
- Butuh kombinasi dengan Session Timeout

---

## ðŸŽ¯ REKOMENDASI WORKFLOW TERBAIK

### Implementasi Bertahap

#### FASE 1: Fix CoA (Priority Tinggi)
1. âœ… Set `default-secret` di `/radius incoming` Mikrotik
2. âœ… Restart Mikrotik
3. âœ… Test dengan `/root/test-coa.sh`
4. âœ… Jika sukses, CoA ready!

#### FASE 2: Tambah Session Timeout (Backup Layer)
1. Update `radgroupreply` untuk include `Session-Timeout`
2. Set timeout = validity voucher (misal 2 jam = 7200 detik)
3. User auto-disconnect setelah timeout (walau CoA gagal)

#### FASE 3: Database Cleanup (Safety Net)
1. Cron mark `acctstoptime` untuk expired vouchers
2. UI filter session yang sudah di-mark
3. Database tetap clean walau Mikrotik belum disconnect

#### FASE 4: Prevention (Long Term)
1. Enable MAC binding untuk voucher
2. Limit concurrent sessions = 1
3. Prevent abuse dan multi-login

---

## ðŸ“‹ Checklist Troubleshooting

Ikuti urutan ini untuk fix CoA:

- [ ] **Step 1:** Verify secret di `/radius incoming` sama dengan `/radius`
- [ ] **Step 2:** Set `default-secret="your-nas-secret"` di `/radius incoming`
- [ ] **Step 3:** Restart Mikrotik (reboot)
- [ ] **Step 4:** Test CoA: `/root/test-coa.sh`
- [ ] **Step 5:** Jika sukses, monitor PM2 logs untuk auto-disconnect
- [ ] **Step 6:** Jika gagal, implement Session Timeout (Opsi A)
- [ ] **Step 7:** Implement database cleanup (Opsi B) sebagai backup
- [ ] **Step 8:** Add MAC binding (Opsi C) untuk prevention

---

## ðŸ” Verifikasi di Mikrotik

Setelah semua setup, verify dengan command berikut:

```routeros
# 1. Cek RADIUS Incoming
/radius incoming print

# Output yang BENAR:
#   accept: yes
#     port: 3799
# default-secret: your-nas-secret

# 2. Cek RADIUS Client
/radius print detail

# Secret harus SAMA dengan incoming!

# 3. Enable logging untuk debug
/system logging
add topics=radius,debug,!packet action=memory

# 4. Test login dengan voucher, tunggu expired, cek log
/log print where topics~"radius"

# Cari log: "received disconnect request" atau "disconnect ack"
```

---

## ðŸ’¡ Kesimpulan

**Prioritas implementasi:**

1. **Fix CoA** (secret + restart) - 95% masalah selesai di sini
2. **Session Timeout** - Backup jika CoA gagal
3. **Database Cleanup** - Safety net untuk UI consistency
4. **MAC Binding** - Prevention untuk abuse

**Jangan pakai Mikrotik API** karena:
- âŒ Overhead untuk cron job (setiap menit)
- âŒ Butuh RouterOS API library
- âŒ Connection pool management kompleks
- âŒ Security risk (harus expose API port)

**Gunakan FreeRADIUS standard CoA** karena:
- âœ… Lightweight (UDP packet)
- âœ… RFC 5176 standard
- âœ… No extra dependencies
- âœ… Supported semua Mikrotik modern
- âœ… Built-in di FreeRADIUS

---

**Next Steps:**
1. Set secret di Mikrotik `/radius incoming`
2. Reboot Mikrotik
3. Test dengan script yang sudah ada
4. Report hasilnya

