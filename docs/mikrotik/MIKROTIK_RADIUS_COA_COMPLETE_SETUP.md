# Panduan Lengkap: Setup Mikrotik RADIUS untuk CoA Disconnect

## ðŸŽ¯ Tujuan
Setup Mikrotik agar dapat menerima CoA (Change of Authorization) Disconnect dari FreeRADIUS server untuk auto-disconnect user expired.

---

## ðŸ“‹ Prerequisites

- Mikrotik RouterOS versi 6.x atau 7.x
- FreeRADIUS server sudah running (VPS: 103.151.140.110)
- Network connectivity antara Mikrotik dan VPS
- Secret RADIUS: gunakan secret dari database (kolom `nas.secret`) (sesuaikan dengan instalasi Anda)

---

## âš™ï¸ Konfigurasi Step-by-Step

### STEP 1: Setup RADIUS Client (Outgoing)

Ini untuk **Mikrotik â†’ FreeRADIUS** (authentication):

```routeros
/radius
add service=hotspot,pppoe \
    address=103.151.140.110 \
    secret="your-nas-secret" \
    timeout=5s
    
# Verify
/radius print detail
```

**Output yang benar:**
```
0   service=hotspot,pppoe 
    address=103.151.140.110 
    secret="your-nas-secret" 
    timeout=5s
```

---

### STEP 2: Setup RADIUS Incoming (CoA)

Ini untuk **FreeRADIUS â†’ Mikrotik** (CoA disconnect):

```routeros
/radius incoming
set accept=yes
set port=3799

# PENTING: Set secret yang SAMA dengan /radius
# Jika tidak ada parameter default-secret, RouterOS versi lama butuh cara lain
print
```

**Output yang benar:**
```
  accept: yes
    port: 3799
```

**âš ï¸ CATATAN PENTING:**

Beberapa versi RouterOS **tidak punya parameter `default-secret`** di `/radius incoming`.

**Solusinya:**
1. Secret untuk incoming CoA **otomatis pakai secret dari `/radius`**
2. Pastikan secret di `/radius` sudah benar
3. **RESTART Mikrotik** agar setting apply

```routeros
# Restart Mikrotik
/system reboot
```

---

### STEP 3: Konfigurasi Firewall

Allow UDP port 3799 dari VPS RADIUS server:

```routeros
/ip firewall filter

# Cek apakah sudah ada rule untuk port 3799
print where dst-port=3799

# Jika belum ada, tambahkan
add chain=input \
    protocol=udp \
    dst-port=3799 \
    src-address=103.151.140.110 \
    action=accept \
    comment="Allow RADIUS CoA from VPS" \
    place-before=0
    
# Verify
print where dst-port=3799
```

**Output yang benar:**
```
0 X  chain=input action=accept protocol=udp src-address=103.151.140.110 
     dst-port=3799 comment="Allow RADIUS CoA from VPS"
```

---

### STEP 4: Enable RADIUS Logging (Debug)

Untuk troubleshooting, enable logging:

```routeros
/system logging

# Tambahkan topic radius
add topics=radius,debug,!packet action=memory

# Verify
print where topics~"radius"
```

---

### STEP 5: Setup Hotspot untuk RADIUS

Pastikan Hotspot Profile menggunakan RADIUS:

```routeros
/ip hotspot profile

# Cek profile yang digunakan
print

# Set RADIUS untuk profile hotspot (misal profile "default")
set [find name=default] use-radius=yes
```

**Verify RADIUS enabled:**
```routeros
/ip hotspot profile print detail
```

Cari baris: `use-radius: yes`

---

## ðŸ§ª Testing

### Test 1: Koneksi Network

Di VPS, test connectivity ke Mikrotik:

```bash
# Test ping
ping -c 3 10.20.30.11

# Test UDP port 3799
nc -vzu 10.20.30.11 3799
```

**Output yang benar:**
```
Connection to 10.20.30.11 3799 port [udp/*] succeeded!
```

---

### Test 2: Manual CoA Disconnect

Di VPS, jalankan test script:

```bash
ssh -p 9500 root@103.151.140.110 /root/test-coa.sh
```

**Output SUKSES:**
```
Received Disconnect-ACK Id 236 from 10.20.30.11:3799
âœ“ CoA command sent successfully
```

**Output GAGAL:**
```
(0) No reply from server for ID 236 socket 3
âœ— CoA command failed
```

Jika GAGAL, lihat section **Troubleshooting** di bawah.

---

### Test 3: Verify di Mikrotik Logs

Setelah kirim CoA, cek log di Mikrotik:

```routeros
/log print where topics~"radius"
```

**Log yang benar (CoA diterima):**
```
received disconnect request from 103.151.140.110
sending disconnect ack to 103.151.140.110
user YFBRLC disconnected
```

**Log yang salah (CoA ditolak):**
```
received disconnect request from 103.151.140.110
bad radius secret from 103.151.140.110
```

Jika secret salah, **secret di `/radius` harus sama dengan database!**

---

### Test 4: Test Auto-Disconnect Voucher

1. **Buat voucher test** (validity 2 menit):
   - Login ke web admin
   - Hotspot â†’ Voucher â†’ Generate
   - Validity: 2 Minute
   - Generate 1 voucher

2. **Login dengan voucher** dari client hotspot

3. **Tunggu 2 menit** sampai expired

4. **Cek PM2 logs** di VPS:
   ```bash
   ssh -p 9500 root@103.151.140.110 "pm2 logs --lines 50 | grep 'CoA\|Disconnect'"
   ```

5. **Verify user disconnect:**
   - Di Mikrotik: `/ip hotspot active print` â†’ voucher hilang
   - Di web: Admin â†’ Sessions â†’ Hotspot â†’ voucher hilang
   - Client redirect ke login page

---

## ðŸ”§ Troubleshooting

### Problem 1: "No reply from server"

**Diagnosa:**
```routeros
# Di Mikrotik, cek log
/log print where topics~"radius"
```

**Kemungkinan:**

1. **Secret tidak match**
   ```routeros
   # Cek secret di RADIUS client
   /radius print detail
   
   # Secret harus SAMA dengan database nas.secret
   ```

2. **Firewall blocking**
   ```routeros
   # Cek firewall
   /ip firewall filter print where dst-port=3799
   
   # Jika tidak ada rule accept, tambahkan
   add chain=input protocol=udp dst-port=3799 src-address=103.151.140.110 action=accept place-before=0
   ```

3. **RADIUS Incoming tidak enabled**
   ```routeros
   /radius incoming print
   
   # Harus: accept: yes, port: 3799
   ```

4. **Butuh restart**
   ```routeros
   # Restart Mikrotik
   /system reboot
   ```

---

### Problem 2: CoA sent tapi user tidak disconnect

**Diagnosa:**

1. **Session ID tidak match**
   ```routeros
   # Cek active session
   /ip hotspot active print
   
   # Bandingkan session-id dengan yang di database
   ```

2. **IP Address tidak match**
   ```routeros
   # Cek IP yang digunakan user
   /ip hotspot active print
   
   # Kolom "address" harus sama dengan framedipaddress di database
   ```

3. **MAC Address strict mode**
   - Beberapa Mikrotik butuh MAC address di CoA packet
   - Update sudah include Calling-Station-Id (MAC)

---

### Problem 3: "Bad RADIUS secret"

**Fix:**

```routeros
# Verify secret di RADIUS client
/radius print detail

# Harus PERSIS SAMA dengan database
# Di database, cek:
# SELECT secret FROM nas WHERE nasname='10.20.30.11'

# Jika beda, update di Mikrotik
/radius set 0 secret="your-nas-secret"

# RESTART
/system reboot
```

---

## ðŸ“Š Monitoring

### Check Active Sessions

```routeros
# Lihat user yang sedang online
/ip hotspot active print

# Detail session
/ip hotspot active print detail

# Lihat berapa lama online
/ip hotspot active print stats
```

### Check RADIUS Statistics

```routeros
# Statistik RADIUS
/radius print stats

# Lihat bad requests (jika ada)
/radius incoming print stats
```

### Check Logs Real-time

```routeros
# Monitor log secara real-time
/log print follow where topics~"radius"
```

**Saat user login:**
```
accepted access request from 103.151.140.110
sending access accept to 103.151.140.110
```

**Saat CoA disconnect:**
```
received disconnect request from 103.151.140.110
sending disconnect ack to 103.151.140.110
```

---

## ðŸŽ¯ Verification Checklist

Pastikan semua ini sudah benar:

- [x] `/radius` configured dengan service=hotspot,pppoe
- [x] Secret di `/radius` = secret di database `nas.secret`
- [x] `/radius incoming` accept=yes port=3799
- [x] Firewall allow UDP 3799 dari VPS
- [x] Hotspot profile use-radius=yes
- [x] Mikrotik sudah restart setelah konfigurasi
- [x] Test CoA manual sukses (Disconnect-ACK received)
- [x] RADIUS logging enabled
- [x] Network connectivity VPS â†” Mikrotik OK

---

## ðŸ’¡ Tips & Best Practices

### 1. Gunakan IP Static untuk VPS

Jangan gunakan dynamic IP untuk RADIUS server. Set di firewall:

```routeros
/ip firewall filter
add chain=input protocol=udp src-address=103.151.140.110 dst-port=1812-1813 action=accept comment="RADIUS Auth"
add chain=input protocol=udp src-address=103.151.140.110 dst-port=3799 action=accept comment="RADIUS CoA"
```

### 2. Backup Konfigurasi

Setelah setup, backup:

```routeros
/export file=radius-config
```

### 3. Monitor Secara Berkala

Cek setiap hari:
- `/log print where topics~"radius" and message~"bad|error|fail"`
- `/radius print stats` â†’ cari bad requests
- `/ip hotspot active print` â†’ pastikan expired user disconnect

### 4. Session Timeout Sebagai Backup

Jika CoA gagal, session tetap disconnect via timeout:

```routeros
# Session akan auto-expire sesuai setting di FreeRADIUS
# Cek di radgroupreply: Session-Timeout attribute
```

---

## ðŸ“š Referensi

- [RFC 5176 - Dynamic Authorization](https://tools.ietf.org/html/rfc5176)
- [Mikrotik Wiki - RADIUS](https://wiki.mikrotik.com/wiki/Manual:RADIUS_Client)
- [FreeRADIUS CoA Documentation](https://networkradius.com/doc/3.0.10/raddb/sites-available/coa.html)

---

## ðŸ†˜ Support

Jika masih bermasalah:

1. Export Mikrotik config: `/export file=debug-config`
2. Cek PM2 logs: `pm2 logs --lines 200 | grep CoA`
3. Run test script: `/root/test-coa.sh`
4. Cek Mikrotik logs: `/log print where topics~"radius"`
5. Screenshot error dan kirim untuk analisis

