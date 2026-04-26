# Mikrotik CoA (Change of Authorization) Setup Guide

## Ringkasan
CoA (RFC 5176) memungkinkan RADIUS server untuk **disconnect user secara paksa** ketika:
- Voucher hotspot expired
- User PPPoE di-isolir
- Admin manual disconnect dari Sesi Hotspot/PPPoE

## Verifikasi Masalah

Jika auto-disconnect tidak bekerja, cek dengan command berikut di VPS:

```bash
# Test CoA manual
echo "NAS-IP-Address=10.20.30.11
Framed-IP-Address=192.168.20.254
User-Name=YFBRLC
Acct-Session-Id=80800031" | radclient -x 10.20.30.11:3799 disconnect your-nas-secret
```

**Error yang umum:**
```
(0) No reply from server for ID 202 socket 3
```

Artinya: **Mikrotik tidak merespon CoA request**

---

## Langkah Setup di Mikrotik

### 1. Enable RADIUS Incoming

```routeros
/radius incoming
set accept=yes
print
```

**Output yang benar:**
```
  accept: yes
    port: 3799
```

### 2. Verifikasi Secret

Secret di Mikrotik RADIUS Client **HARUS SAMA** dengan secret di database `nas` table.

Cek di aplikasi web: **Admin â†’ Network â†’ Routers** â†’ pilih router â†’ lihat field "Secret"

Atau cek via WinBox/SSH di Mikrotik:
```routeros
/radius
print detail
```

Contoh output:
```
0   service=hotspot,pppoe address=103.151.140.110 secret="your-nas-secret"
    timeout=5s00ms
```

**PENTING:** Secret harus **PERSIS SAMA** antara:
- Database `nas.secret`
- Mikrotik `/radius` secret
- Test script CoA

### 3. Enable RADIUS Logging (Untuk Debug)

```routeros
/system logging
add topics=radius,debug,!packet action=memory

# Lihat log
/log print where topics~"radius"
```

### 4. Verifikasi Firewall

Pastikan UDP port 3799 **TIDAK DIBLOK** oleh firewall Mikrotik:

```routeros
# Cek apakah ada rule yang block port 3799
/ip firewall filter
print where dst-port=3799

# Jika tidak ada rule ACCEPT, tambahkan
add chain=input protocol=udp dst-port=3799 \
    src-address=103.151.140.110 \
    action=accept \
    comment="Allow RADIUS CoA from VPS" \
    place-before=0
```

**Catatan:** `103.151.140.110` adalah IP VPS RADIUS server Anda.

### 5. Test Koneksi dari VPS

Di VPS, jalankan:

```bash
# Test ping
ping -c 3 10.20.30.11

# Test UDP port (harus succeed)
nc -vzu 10.20.30.11 3799
```

**Output yang benar:**
```
Connection to 10.20.30.11 3799 port [udp/*] succeeded!
```

---

## Troubleshooting

### Problem: "No reply from server"

**Penyebab Umum:**

1. **RADIUS Incoming tidak enabled**
   ```routeros
   /radius incoming set accept=yes
   ```

2. **Secret tidak match**
   - Cek di Mikrotik: `/radius print detail`
   - Cek di database: `SELECT secret FROM nas WHERE nasname='10.20.30.11'`
   - Harus **IDENTIK** (case-sensitive)

3. **Firewall blocking UDP 3799**
   ```routeros
   /ip firewall filter
   add chain=input protocol=udp dst-port=3799 action=accept place-before=0
   ```

4. **Network routing issue**
   - Pastikan VPS bisa ping ke IP Mikrotik
   - Pastikan tidak ada firewall di tengah (VPN, NAT, dll)

### Problem: CoA sent tapi user tidak disconnect

1. **Session ID tidak match**
   - CoA harus menggunakan `acctsessionid` yang **AKTIF**
   - Cek di Mikrotik: `/ip hotspot active print`
   - Cek di database: `SELECT acctsessionid FROM radacct WHERE acctstoptime IS NULL`

2. **Framed-IP tidak match**
   - Harus pakai IP yang **sedang digunakan** oleh user
   - Cek di Mikrotik: `/ip hotspot active print` â†’ kolom "address"

3. **NAS-IP-Address salah**
   - Harus pakai IP Mikrotik yang **menerima** RADIUS request
   - Biasanya sama dengan `nasipaddress` di tabel `radacct`

---

## Test Script

Gunakan script `/root/test-coa.sh` di VPS untuk test manual:

```bash
ssh -p 9500 root@103.151.140.110 /root/test-coa.sh
```

**Output sukses:**
```
Received Disconnect-ACK Id 236 from 10.20.30.11:3799 to 0.0.0.0:54877 length 20
âœ“ CoA command sent successfully
```

**Output gagal:**
```
(0) No reply from server for ID 236 socket 3
âœ— CoA command failed
```

---

## Verifikasi Auto-Disconnect Bekerja

### 1. Buat voucher test (validity 2 menit)

Di web admin:
- Admin â†’ Hotspot â†’ Voucher â†’ Generate
- Profile: pilih profile hotspot
- Validity: 2 Minute
- Quantity: 1

### 2. Login dengan voucher

Dari client hotspot, login dengan voucher yang baru dibuat.

### 3. Tunggu sampai expired

Setelah 2 menit, voucher akan expired. Cek PM2 logs di VPS:

```bash
ssh -p 9500 root@103.151.140.110 "pm2 logs --lines 50 | grep 'CoA\|Disconnect'"
```

**Output yang benar:**
```
[CRON] Found active session for EXPIRED voucher XXXXXX - disconnecting
[CoA] Disconnecting XXXXXX from 10.20.30.11:3799
[CoA] âœ“ Successfully sent disconnect for XXXXXX
[CRON] âœ“ Disconnected XXXXXX from test (10.20.30.11)
```

### 4. Verifikasi user ter-disconnect

- Cek di Mikrotik: `/ip hotspot active print` â†’ voucher sudah hilang
- Cek di web: Admin â†’ Sessions â†’ Hotspot â†’ voucher sudah hilang
- Client akan otomatis redirect ke login page

---

## Catatan Penting

### Kenapa CoA Penting?

Tanpa CoA, user yang voucher-nya sudah expired bisa **tetap online** sampai:
- Koneksi terputus manual (reboot, logout)
- Session timeout dari Mikrotik
- Network disconnect

Dengan CoA, user **langsung ter-kick** begitu voucher expired.

### Alternatif Jika CoA Tidak Bisa Dipakai

Jika Mikrotik di belakang NAT/firewall yang tidak bisa dibuka port 3799:

1. **Session Timeout di Hotspot Profile**
   - Set `Session Timeout` di Mikrotik Hotspot Profile (misal 1 jam)
   - User akan auto-disconnect setelah timeout
   - **Kekurangan:** Tidak instant, masih bisa online beberapa lama

2. **Idle Timeout**
   - Set `Idle Timeout` di Hotspot Profile
   - User disconnect jika idle (tidak ada traffic)
   - **Kekurangan:** User yang aktif tetap online walau expired

3. **Manual Disconnect via Mikrotik API**
   - Aplikasi connect ke Mikrotik API dan hapus session manual
   - **Kekurangan:** Butuh Mikrotik API enabled, lebih kompleks

**Rekomendasi:** Selalu gunakan CoA untuk hasil terbaik.

---

## Support

Jika masih bermasalah setelah mengikuti guide ini:

1. Export Mikrotik config dan cek bagian `/radius` dan `/radius incoming`
2. Cek PM2 logs: `pm2 logs --lines 200 | grep CoA`
3. Test manual dengan script `/root/test-coa.sh`
4. Screenshoot error dan log untuk troubleshooting lebih lanjut

