# SALFANET RADIUS - Billing System for ISP/RTRW.NET

Modern, full-stack billing & RADIUS management system for ISP/RTRW.NET with FreeRADIUS integration supporting PPPoE and Hotspot authentication.

> **Latest:** v2.25.0 — Build APK Android langsung di server VPS, download APK tanpa GitHub Actions (Apr 26, 2026)

---

## 🤖 AI Development Assistant

**READ FIRST:** [docs/AI_PROJECT_MEMORY.md](docs/AI_PROJECT_MEMORY.md) — contains full architecture, VPS details, DB schema, known issues, and proven solutions.

---

## 🎯 Features

| Category | Key Capabilities |
|----------|-----------------|
| **RADIUS / Auth** | FreeRADIUS 3.0.26, PAP/CHAP/MS-CHAP, VPN L2TP/IPSec, PPPoE & Hotspot, CoA real-time speed/disconnect |
| **VPN Management** | MikroTik CHR via API, VPS built-in WireGuard & L2TP/IPsec peer management, configurable IP pool & gateway per protocol, auto-generated RouterOS scripts |
| **PPPoE Management** | Customer accounts, profile-based bandwidth, isolation, IP assignment, MikroTik auto-sync, foto KTP+instalasi via kamera HP, GPS otomatis |
| **Hotspot Voucher** | 8 code types, batch up to 25,000, agent distribution, auto-sync with RADIUS, print templates |
| **Billing** | Postpaid/prepaid invoices, auto-generation, payment reminders, balance/deposit, auto-renewal |
| **Payment** | Manual upload (bukti transfer), Midtrans/Xendit/Duitku gateway, approval workflow, 0–5 bank accounts |
| **Notifications** | WhatsApp (Fonnte/WAHA/GOWA/MPWA/Wablas/WABlast/**Kirimi.id**), Email SMTP, broadcast (outage/invoice/payment), webhook pesan masuk |
| **Agent/Reseller** | Balance-based voucher generation, commission tracking, sales stats |
| **Financial** | Income/expense tracking with categories, keuangan reconciliation |
| **Network (FTTH)** | OLT/ODC/ODP management, customer port assignment, network map, distance calculation |
| **GenieACS TR-069** | CPE/ONT management, WiFi config (SSID/password), device status & uptime |
| **Isolation** | Auto-isolate expired customers, customizable WhatsApp/Email/HTML landing page templates |
| **Cron Jobs** | 16 automated background jobs (tsx runner via PM2 fork), history, distributed locking, manual trigger |
| **Roles & Permissions** | 53 permissions, 5 portals (Admin/Customer/Agent/Technician + SuperAdmin) |
| **Activity Log** | Audit trail with auto-cleanup (30 days) |
| **Security** | Session timeout 30 min, idle warning, RBAC, HTTPS/SSL |
| **Bahasa** | Bahasa Indonesia (full) |
| **PWA** | Installable di semua portal (admin, customer, agent, technician), offline fallback, service worker cache |
| **Web Push** | VAPID-based browser push notifications, subscribe/unsubscribe toggle, admin broadcast |
| **System Update** | Update via SSH menggunakan `updater.sh`, tidak ada web-based update |
| **Mobile App** | Flutter customer portal (WiFi control, invoice, payment) |
| **Android APK Builder** | Build APK Android (WebView wrapper) langsung di server VPS untuk 4 portal (Admin/Customer/Technician/Agent), download APK tanpa GitHub Actions |

---

## 🚀 Tech Stack

| Component | Technology |
|-----------|------------|
| Framework | Next.js 16 (App Router, standalone output) |
| Language | TypeScript |
| Styling | Tailwind CSS |
| Database | MySQL 8.0 + Prisma ORM |
| RADIUS | FreeRADIUS 3.0.26 |
| Process Manager | PM2 (cluster × 2) |
| Session Tracking | FreeRADIUS radacct (real-time) |
| Maps | Leaflet / OpenStreetMap |

---

## 📁 Project Structure

```
salfanet-radius/
├── src/
│   ├── app/
│   │   ├── admin/          # Admin panel
│   │   ├── agent/          # Agent/reseller portal
│   │   ├── api/            # API route handlers
│   │   ├── customer/       # Customer self-service portal
│   │   └── technician/     # Technician portal
│   ├── server/             # DB, services, jobs, cache, auth
│   ├── features/           # Vertical slices (queries, schemas, types)
│   ├── components/         # Shared React components
│   ├── locales/            # i18n translations (id, en)
│   └── types/              # Shared TypeScript types
├── prisma/
│   ├── schema.prisma       # Database schema (~45 models)
│   └── seeds/              # Seed scripts
├── freeradius-config/      # FreeRADIUS config (deployed by installer)
├── vps-install/            # One-command VPS installer scripts
├── production/             # PM2 & Nginx config templates
├── mobile-app/             # Flutter customer app
├── scripts/                # Utility & tuning scripts
└── docs/                   # Documentation & AI memory
```

---

## ⚙️ Installation

### Metode 1 — Git Clone (Recommended)

```bash
ssh root@YOUR_VPS_IP

git clone https://github.com/s4lfanet/salfanet-radius.git /root/salfanet-radius
cd /root/salfanet-radius
bash vps-install/vps-installer.sh
```

Installer akan berjalan **interaktif** — mendeteksi environment otomatis, memandu konfigurasi, lalu menjalankan semua step.

---

### Metode 2 — Upload Manual via SCP (Tanpa Akses Internet di Server)

```bash
# Jalankan di terminal LOKAL (bukan di server)
scp -r ./salfanet-radius root@YOUR_VPS_IP:/root/salfanet-radius

# SSH ke server, lalu jalankan installer
ssh root@YOUR_VPS_IP
cd /root/salfanet-radius
bash vps-install/vps-installer.sh
```

---

### Environment yang Didukung

| Environment | Flag | Akses |
|------------|------|-------|
| **Public VPS** (DigitalOcean, Vultr, Hetzner, AWS) | `--env vps` | Internet |
| **Proxmox LXC** | `--env lxc` | LAN/VLAN |
| **Proxmox VM / VirtualBox** | `--env vm` | LAN |
| **Bare Metal / Server Fisik** | `--env bare` | LAN |

```bash
# Contoh: paksa environment + IP
bash vps-install/vps-installer.sh --env lxc --ip 192.168.1.50
```

---

### Updating Existing Installation

Cara paling aman. **Semua data upload (logo, foto KTP pelanggan, bukti bayar) otomatis dipreservasi.**

```bash
bash /var/www/salfanet-radius/vps-install/updater.sh
```

Atau update dari branch terbaru secara manual:

```bash
cd /var/www/salfanet-radius
git pull origin master
npm install --legacy-peer-deps
npx prisma db push
npm run build
pm2 reload all
```

Lihat detail lengkap di [vps-install/README.md](vps-install/README.md).

---

### Data yang Aman Saat Update

| Data | Status |
|------|--------|
| Logo perusahaan (`public/uploads/logos/`) | ✅ Dipreservasi |
| Foto KTP & dokumen pelanggan | ✅ Dipreservasi |
| Bukti pembayaran | ✅ Dipreservasi |
| File `.env` (database, secrets) | ✅ Tidak disentuh |
| **Database MySQL (semua data pelanggan)** | ✅ Tidak disentuh |

---

### Default Credentials

| | |
|--|--|
| Admin URL | `http://YOUR_VPS_IP/admin/login` |
| Username | `superadmin` |
| Password | `admin123` |

⚠️ **Ganti password segera setelah login pertama!**

---

## 🔌 FreeRADIUS

Key config files at `/etc/freeradius/3.0/`:

| File | Purpose |
|------|---------|
| `mods-enabled/sql` | MySQL connection for user auth |
| `mods-enabled/rest` | REST API for voucher management |
| `sites-enabled/default` | Main auth logic (PPPoE realm support) |
| `clients.conf` | NAS/router clients (+ `$INCLUDE clients.d/`) |
| `sites-enabled/coa` | CoA/Disconnect-Request virtual server |

Config backup in `freeradius-config/` is auto-deployed by the installer.

### Auth Flow

**PPPoE:** `MikroTik → FreeRADIUS → MySQL (radcheck/radusergroup/radgroupreply)` → Access-Accept with Mikrotik-Rate-Limit

**Hotspot Voucher:** Same RADIUS path + `REST /api/radius/post-auth` → sets firstLoginAt, expiresAt, syncs keuangan

### RADIUS Tables

| Table | Purpose |
|-------|---------|
| `radcheck` | User credentials |
| `radreply` | User-specific reply attrs |
| `radusergroup` | User → Group mapping |
| `radgroupreply` | Group reply (bandwidth, session timeout) |
| `radacct` | Session accounting |
| `nas` | NAS/Router clients (dynamic) |

---

## ⏰ Cron Jobs (16 automated)

| Job | Schedule | Function |
|-----|----------|----------|
| Voucher Sync | Every 5 min | Sync voucher status with RADIUS |
| Disconnect Sessions | Every 5 min | CoA disconnect expired vouchers |
| Auto Isolir (PPPoE) | Every hour | Suspend overdue customers |
| FreeRADIUS Health | Every 5 min | Auto-restart if down |
| PPPoE Session Sync | Every 10 min | Sync radacct sessions |
| Agent Sales | Daily 1 AM | Update sales statistics |
| Invoice Generate | Daily 2 AM | Generate monthly invoices |
| Activity Log Cleanup | Daily 2 AM | Delete logs >30 days |
| Invoice Reminder | Daily 8 AM | Send payment reminders |
| Invoice Status | Daily 9 AM | Mark overdue invoices |
| Notification Check | Every 10 min | Process notification queue |
| Auto Renewal | Daily 8 AM | Prepaid auto-renew from balance |
| Webhook Log Cleanup | Daily 3 AM | Delete webhook logs >30 days |
| Session Monitor | Every 5 min | Security session monitoring |
| Cron History Cleanup | Daily 4 AM | Keep last 50 per job type |
| Suspend Check | Every hour | Activate/restore suspend requests |

All jobs can be triggered manually from **Settings → Cron** in the admin panel.

---

## � Android APK Builder

Buat APK Android (WebView wrapper) untuk 4 portal langsung di server VPS — tanpa GitHub Actions, tanpa Android Studio.

### 1) Setup Android SDK (satu kali via SSH)

```bash
apt-get update && apt-get install -y openjdk-17-jdk wget unzip && \
mkdir -p /opt/android/cmdline-tools && \
wget -q https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip -O /tmp/cmdtools.zip && \
unzip -q /tmp/cmdtools.zip -d /opt/android/cmdline-tools && \
mv /opt/android/cmdline-tools/cmdline-tools /opt/android/cmdline-tools/latest && \
yes | /opt/android/cmdline-tools/latest/bin/sdkmanager --licenses && \
/opt/android/cmdline-tools/latest/bin/sdkmanager "platforms;android-34" "build-tools;34.0.0" && \
echo 'export ANDROID_HOME=/opt/android' >> /etc/environment && \
echo 'Selesai!'
```

> **Perkiraan waktu:** ~5–10 menit (download ~500MB). Disk yang dibutuhkan: ~2GB.

### 2) Build APK via Admin Panel

Buka **Admin → Download Aplikasi Android** → klik **Build APK** pada role yang diinginkan.

- Build berjalan di background (tidak timeout meski butuh beberapa menit)
- Status diperbarui otomatis setiap 3 detik
- Setelah selesai, tombol **Download APK** muncul

### 3) Build via API (opsional)

```bash
# Cek environment
curl http://YOUR_VPS/api/admin/apk/trigger

# Mulai build (role: admin | customer | technician | agent)
curl -X POST http://YOUR_VPS/api/admin/apk/trigger?role=customer \
  -H "Cookie: next-auth.session-token=..."

# Cek status
curl http://YOUR_VPS/api/admin/apk/status?role=customer

# Download APK
curl -OJ http://YOUR_VPS/api/admin/apk/file?role=customer \
  -H "Cookie: next-auth.session-token=..."
```

### Storage APK

| Path | Keterangan |
|------|------------|
| `/var/data/salfanet/apk/{role}/app.apk` | File APK hasil build |
| `/var/data/salfanet/apk/{role}/status.json` | Status & metadata build |
| `/var/data/salfanet/apk/{role}/build.log` | Log Gradle |
| `/var/data/salfanet/gradle-cache` | Cache Gradle (mempercepat build berikutnya) |

### Paket Aplikasi

| Role | Package ID | Warna |
|------|-----------|-------|
| Admin | `net.salfanet.admin` | Biru |
| Customer | `net.salfanet.customer` | Cyan |
| Technician | `net.salfanet.technician` | Hijau |
| Agent | `net.salfanet.agent` | Ungu |

---

## �🛠️ Common Commands

```bash
# PM2
pm2 status ; pm2 logs salfanet-radius
pm2 restart ecosystem.config.js --update-env

# FreeRADIUS
systemctl restart freeradius
freeradius -XC    # Test config
radtest 'user@realm' password 127.0.0.1 0 testing123

# Database
mysql -u salfanet_user -psalfanetradius123 salfanet_radius
mysqldump -u salfanet_user -psalfanetradius123 salfanet_radius > backup.sql
```

---

## 🧯 Troubleshooting Cepat

### 1) Website tidak bisa diakses dari IP VPS

Jika `Nginx` dan app sudah jalan di server tapi dari internet tetap tidak bisa akses, biasanya masalah ada di layer jaringan (NAT/forwarding/firewall external), bukan di aplikasi.

```bash
# Di VM/VPS guest
ss -tulpn | grep -E ':80|:443|:3000'
curl -I http://127.0.0.1:3000
curl -I http://127.0.0.1
systemctl status nginx --no-pager
pm2 status
```

Jika semua check local di atas OK, cek mapping di host Proxmox/router/cloud firewall:

1. `Public:2020 -> VM:22` (SSH)
2. `Public:80 -> VM:80` (HTTP)
3. `Public:443 -> VM:443` (HTTPS)

Catatan: `IP:2020` adalah port SSH, bukan URL web aplikasi.

### 2) PM2 jalan tapi web tetap blank/error

```bash
pm2 status
pm2 logs salfanet-radius --lines 100
cd /var/www/salfanet-radius
npm run build
pm2 restart ecosystem.config.js --update-env
```

### 4) Jalankan diagnosa Nginx otomatis dari installer

Installer Nginx terbaru menambahkan self-check internal (`127.0.0.1:3000`, `127.0.0.1`) dan best-effort check publik (HTTP/HTTPS).

```bash
cd /var/www/salfanet-radius
bash vps-install/install-nginx.sh
```

Jika warning menunjukkan HTTP publik tidak reachable, fokus perbaikan di NAT/port-forward/security-group, bukan di Next.js.

---

## 🔐 Security

```bash
# Firewall
ufw allow 22/tcp && ufw allow 80/tcp && ufw allow 443/tcp
ufw allow 1812/udp && ufw allow 1813/udp && ufw allow 3799/udp
```

1. Change default admin password on first login
2. Change MySQL passwords in `.env`
3. Configure SSL (Let's Encrypt or Cloudflare)
4. Enable UFW

---

## 📡 CoA (Change of Authorization)

Sends real-time speed/disconnect commands to MikroTik without dropping PPPoE connections.

**MikroTik requirement:** `/radius incoming set accept=yes port=3799`

**API:** `POST /api/radius/coa` — actions: `disconnect`, `update`, `sync-profile`, `test`

Auto-triggered when: PPPoE profile speed is edited (syncs all active sessions).

---

## 📲 WhatsApp Providers

| Provider | Base URL | Auth |
|----------|----------|------|
| Fonnte | `https://api.fonnte.com/send` | Token |
| WAHA | `http://IP:PORT` | API Key |
| GOWA | `http://IP:PORT` | `user:pass` |
| MPWA | `http://IP:PORT` | API Key |
| Wablas | `https://pati.wablas.com` | Token |

---

## ⏱️ Timezone

| Layer | Timezone | Note |
|-------|----------|------|
| Database (Prisma) | UTC | Prisma default |
| FreeRADIUS | WIB (UTC+7) | Server local time |
| PM2 env | WIB | `TZ: 'Asia/Jakarta'` in ecosystem.config.js |
| API / Frontend | WIB | Auto-converts UTC ↔ WIB |

For WITA (UTC+8) or WIT (UTC+9): change `TZ` in `.env`, `ecosystem.config.js`, and `src/lib/timezone.ts`.

---

## 📋 Admin Modules

Dashboard · PPPoE · Hotspot · Agent · Invoice · Payment · Keuangan · Sessions · WhatsApp · Network (OLT/ODC/ODP) · GenieACS · Settings

**Roles:** SUPER_ADMIN · FINANCE · CUSTOMER_SERVICE · TECHNICIAN · MARKETING · VIEWER

---

## 📝 Changelog

Bagian ini otomatis sinkron dari `CHANGELOG.md` saat file changelog berubah di GitHub.

<!-- AUTO-CHANGELOG:START -->

### v2.25.1 — 2026-04-26

### Added
- **`vps-install/install-security.sh` — Modul keamanan server otomatis** — Script baru yang dipanggil di Step 8 installer dan setiap `updater.sh`. Memasang tiga lapisan perlindungan secara otomatis:
  - **fail2ban**: ban IP brute-force SSH setelah 5x gagal dalam 10 menit (ban 2 jam). Jail aktif: `sshd`, `nginx-http-auth`, `nginx-limit-req`. IP jaringan lokal (`192.168.x.x`, `10.x.x.x`) tidak pernah di-ban.
  - **UFW Firewall**: default deny semua incoming, allow hanya port yang dibutuhkan: 22/TCP (SSH), 80/TCP (HTTP), 443/TCP (HTTPS), 1812-1813/UDP (RADIUS), 3799/UDP (RADIUS CoA). Di-skip otomatis untuk LXC container (pakai Proxmox host firewall).
  - **Disk cleanup cronjob**: script `/usr/local/bin/salfanet-cleanup.sh` berjalan otomatis setiap hari jam 02:00. Membersihkan: journal systemd (max 200MB/7 hari), syslog lama, btmp (truncate jika >50MB), APT cache, tmp files, PM2 logs besar, Gradle cache >30 hari, APK build temp.
  - Bisa dijalankan manual: `bash vps-install/install-security.sh`
  - Log cleanup: `/var/log/salfanet-cleanup.log` (auto-trim jika >5MB)

### Fixed
- **Disk penuh 100% menyebabkan MySQL deadlock & API 500** — Disk VPS publik penuh akibat log systemd journal (~2.9GB) dan syslog (~2.2GB) menumpuk. MySQL tidak bisa commit karena disk penuh → semua query FreeRADIUS (`radpostauth`, `radacct`) stuck "waiting for handler commit" → Prisma connection pool exhausted (P2024) → semua API endpoint 500. Diatasi dengan cleanup log + install cronjob harian.
- **Build APK customer/technician/agent: connection pool exhausted saat 3 build serentak** — Menjalankan Gradle build untuk 3 role sekaligus menyebabkan VPS overload. Prisma connection pool (limit 10) habis karena server tidak bisa melayani request DB selama build berjalan. Build sebenarnya tetap berjalan di background; yang "berhenti" hanya tampilan UI karena polling API gagal 500.

### Changed
- **`vps-install/vps-installer.sh`: tambah Step 8 (Security)** — Installer utama kini memanggil `install-security.sh` secara otomatis setelah Step 7 (PM2 & Build). Instalasi baru langsung terlindungi fail2ban + UFW + cleanup cron tanpa langkah manual.
- **`vps-install/updater.sh`: security check saat setiap update** — Setiap kali `bash updater.sh` dijalankan, script memastikan: (1) cleanup cronjob terpasang, (2) fail2ban dalam keadaan running. Idempotent — aman dijalankan berulang kali.

---

### v2.25.0 — 2026-04-26

### Added
- **Build APK Android langsung di server VPS** ([`91a45d5`]) — Fitur baru di halaman `/admin/download-apk`: build APK Android Kotlin (WebView wrapper) langsung di server menggunakan Gradle, tanpa perlu upload ke GitHub atau install Android Studio. APK tersimpan di server dan bisa didownload kapan saja.
  - `GET /api/admin/apk/trigger` — cek ketersediaan Java JDK dan Android SDK di server
  - `POST /api/admin/apk/trigger?role=admin|customer|technician|agent` — mulai build di background (detached process, tidak timeout)
  - `GET /api/admin/apk/status?role=...` — polling status build: `idle` / `building` / `done` / `failed` / `stale`
  - `GET /api/admin/apk/file?role=...` — download APK hasil build
  - UI polling otomatis setiap 3 detik selama build berjalan
  - Deteksi stale build: jika status masih `building` setelah 15 menit, otomatis ditandai `stale`
  - Panduan install Android SDK ditampilkan di UI jika environment belum siap (copy-able bash command)
  - Fallback ZIP download tetap tersedia via collapsible section
- **Setup Android SDK di VPS** (manual, satu kali) — Jalankan command berikut via SSH sebelum menggunakan fitur build:
  ```bash
  apt-get update && apt-get install -y openjdk-17-jdk wget unzip && \
  mkdir -p /opt/android/cmdline-tools && \
  wget -q https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip -O /tmp/cmdtools.zip && \
  unzip -q /tmp/cmdtools.zip -d /opt/android/cmdline-tools && \
  mv /opt/android/cmdline-tools/cmdline-tools /opt/android/cmdline-tools/latest && \
  yes | /opt/android/cmdline-tools/latest/bin/sdkmanager --licenses && \
  /opt/android/cmdline-tools/latest/bin/sdkmanager "platforms;android-34" "build-tools;34.0.0" && \
  echo 'export ANDROID_HOME=/opt/android' >> /etc/environment && \
  echo 'Selesai!'
  ```
  Build pertama ±3–5 menit (download Gradle dependencies). Build berikutnya ±1 menit (Gradle cache di `/var/data/salfanet/gradle-cache`).

---

### v2.24.0 — 2026-04-26

### Removed
- **Update otomatis via web panel dihapus** ([`4692059`]) — Fitur update via browser (SSE live log, tombol Apply Update / Force Rebuild) dihapus karena tidak reliable: bash script `update.sh` selalu mati saat `pm2 stop` dipanggil dari dalam Next.js process group, menyebabkan `.next` terhapus dan server 502 yang harus dipulihkan manual. File yang dihapus:
  - `scripts/update.sh` — script update yang dipanggil via API
  - `src/app/api/admin/system/update/route.ts` — SSE API endpoint (GET stream + POST trigger)
  - Halaman `/admin/system` diganti menjadi halaman **Informasi Sistem** statis: versi, commit, Node.js, uptime, banner update tersedia, dan panduan SSH siap-copy untuk update manual

### Fixed
- **`vps-install/updater.sh`: default ke `--branch master`** ([`5aa05b7`]) — Menjalankan `bash updater.sh` tanpa flag sebelumnya masuk ke Mode B (GitHub Releases) yang langsung error 404 karena repo tidak menggunakan GitHub Releases. Sekarang jika tidak ada `--branch` maupun `--version`, script otomatis pakai `--branch master`.

---

### v2.23.0 — 2026-04-26

### Removed
- **Coordinator role dihapus sepenuhnya** ([`e0cd701`]) — Role coordinator adalah fitur yang tidak pernah selesai diimplementasi. Semua endpoint API tidak pernah dibuat, sehingga halaman-halamannya selalu error. File yang dihapus:
  - `src/app/coordinator/` — seluruh direktori portal coordinator (dashboard, tasks)
  - `src/app/admin/coordinators/` — halaman manajemen coordinator di admin panel
  - `src/locales/id.json` — key `coordinator`, `coordinatorLogin`, `manageCoordinators`, namespace `"coordinator"` (~40 key), dan `"senderType_COORDINATOR"` dihapus
  - `src/app/admin/tickets/[id]/page.tsx` — `COORDINATOR` dihapus dari `SenderType` union type dan dari objek styling `getSenderBadgeColor()`
- **Firebase Admin SDK & FCM dihapus** ([`fdc730b`]) — Seluruh integrasi Firebase Cloud Messaging dihapus. Push notification kini menggunakan VAPID Web Push murni (tidak ada dependency firebase-admin). File yang dihapus: `src/server/push.service.ts`, `firebase-service-account.json`. Stub `firebase-admin` di `src/lib/` digantikan dengan implementasi VAPID native.

### Added
- **`src/cron/runner.ts` — Cron runner baru berbasis tsx** ([`fdc730b`]) — Menggantikan `cron-service.js` (Node.js CJS) dengan TypeScript runner yang dijalankan via `npx tsx`. 16 cron jobs diload dari satu entry point, distributed locking tetap aktif. FreeRADIUS Health Check berjalan 5 detik setelah startup.
- **`production/ecosystem.config.js` — Template konfigurasi PM2** ([`fdc730b`]) — File baru sebagai source of truth untuk konfigurasi PM2. `salfanet-cron` kini berjalan sebagai proses fork (`npx tsx src/cron/runner.ts`) dengan `NODE_OPTIONS: '--conditions=react-server'` (wajib agar `server-only` package tidak throw di luar Next.js).
- **`vps-install/cleanup-refactor.sh` — Script cleanup instalasi lama** ([`f71256c`], [`c41f44f`]) — Script idempotent untuk membersihkan file-file stale dari instalasi sebelum refactor. Fitur:
  - Support `--dry-run` (preview tanpa hapus)
  - Phase 1: cleanup Firebase/FCM push service, firebase-service-account.json
  - Phase 3: sync `ecosystem.config.js` dari `production/` (migrasi cron-service.js → tsx runner)
  - Phase 8: hapus `src/app/coordinator/`, `src/app/admin/coordinators/`
  - Auto-deteksi jika `salfanet-cron` masih pakai `cron-service.js` → migrate ke tsx runner otomatis
  - Usage: `bash vps-install/cleanup-refactor.sh [--dry-run] [--app-dir=/path]`

### Changed
- **`scripts/update.sh`: refactor-aware** ([`f71256c`]) — Update script (dipanggil via admin panel → `/api/admin/system/update`) ditingkatkan:
  - Setelah `git reset --hard`, otomatis copy `production/ecosystem.config.js` → root (file ini untracked, tidak tereset oleh git)
  - Cleanup stale files dari Phase 1-8 refactor (push.service.ts, coordinator, firebase, dll.) di setiap update
  - PM2 cron restart: jika `ecosystem.config.js` berubah → `pm2 delete` + `pm2 start` ulang (bukan sekedar `pm2 restart`)
  - `pm2 save` otomatis setelah restart
- **`vps-install/updater.sh`: refactor-aware** ([`f71256c`]) — CLI update script ditingkatkan:
  - `npm ci` dengan fallback ke `npm install --production=false` jika lock file tidak sinkron (umum terjadi setelah refactor)
  - Copy `production/ecosystem.config.js` setelah `git clean -fd`
  - Cleanup stale files refactor (list sama dengan update.sh)
  - Copy static assets ke `.next/standalone` setelah build
  - PM2 cron: deteksi perubahan script → `pm2 delete` + `pm2 start` jika perlu

### Fixed
- **`cleanup-refactor.sh`: `set -e` safe** ([`c41f44f`]) — Fungsi `remove_path()` sebelumnya `return 1` saat file tidak ditemukan → script keluar prematur karena `set -e`. Diperbaiki ke `return 0`. Kondisi `diff` juga diperbaiki (inversi `!` yang salah menyebabkan ecosystem.config.js tidak pernah disync).

---

### v2.22.0 — 2026-04-26

### Added
- **Script `scripts/backup-freeradius-local.sh`** ([`8652ea4`]) — Script bash untuk membuat arsip `.tar.gz` seluruh direktori `/etc/freeradius/3.0/` ke `backups/freeradius/` dengan nama file bertimestamp (`freeradius-config-YYYYMMDD-HHMMSS.tar.gz`). Otomatis cleanup backup lama (simpan 10 terbaru). Output baris `BACKUP_FILE: <nama>` di akhir agar UI polling bisa deteksi selesai. Script sebelumnya tidak ada sehingga tombol "Buat Backup" selalu gagal dengan error `Script not found`.

### Fixed
- **Restore FreeRADIUS: error "same file" saat restore `mods-enabled/`** ([`c268123`]) — File `mods-enabled/sql` dan `mods-enabled/rest` di FreeRADIUS adalah **symlink** ke `../mods-available/sql`. Saat tar mengekstrak backup, symlink tetap sebagai symlink. Perintah `cp symlink dest` gagal karena keduanya resolve ke file fisik yang sama (`cp: ... are the same file`). Fix: cek tipe file via `stat -c '%F'` sebelum restore — jika `symbolic link`, gunakan `ln -sf <target> <dest>` alih-alih `cp`.
- **Build VPS: OOM (Out of Memory) saat fase TypeScript check** ([`0aee02f`]) — Build `npm run build` menjalankan TypeScript type-checker (`tsc`) setelah compile selesai. Pada VPS 4GB dengan PM2 berjalan, proses `tsc` membutuhkan heap hingga 1.6GB dan di-kill oleh OOM killer (`FATAL ERROR: Ineffective mark-compacts near heap limit`). Fix: set `typescript.ignoreBuildErrors: true` di `next.config.ts` untuk skip fase `tsc` saat build produksi (type error tetap terdeteksi di development/editor).
- **Build VPS: OOM saat build karena PM2 mengonsumsi RAM** ([`08eba82`]) — PM2 process salfanet-radius mengonsumsi ~500MB RAM saat berjalan. Dengan heap build 1536MB (bawaan `npm run build`), total RAM yang dibutuhkan melebihi 4GB. Fix: `update.sh` kini stop PM2 sebelum build dan gunakan `npm run build:low-mem` (heap 1024MB). PM2 distart kembali setelah build selesai (atau gagal).
- **Build VPS: script baru tidak executable setelah `git reset --hard`** ([`8ce6421`]) — Script yang ditambahkan via commit baru tidak otomatis dapat izin execute di VPS setelah `git reset --hard`. Fix: tambah `chmod +x scripts/*.sh` di `update.sh` setelah git reset.
- **VPN Client: list tidak refresh setelah tambah client** ([`b55d3e6`]) — Setelah berhasil tambah WireGuard atau L2TP client, list VPN tidak diperbarui otomatis. Fix: panggil `loadClients()` di success path WireGuard dan L2TP.
- **VPN Client: modal tidak menutup / formData tidak ter-reset setelah submit** ([`b55d3e6`]) — Form WireGuard menggunakan `formData.name` setelah `formData` di-clear sehingga nama yang dikirim ke credentials dialog kosong. Fix: simpan nama ke variabel lokal `peerName` sebelum clear, gunakan `peerName` di credentials dialog.
- **VPN Client: IP pool tidak bisa dipakai ulang (orphan WG peers)** ([`288a094`]) — Peer WireGuard yang dihapus dari DB tetap tersisa di `wg.conf`. Saat tambah client baru, `nextAvailableIp` membaca `wg.conf` dan skip IP yang sebenarnya sudah bebas. Fix: tambah langkah cleanup orphan peers di `wg.conf` (compare dengan DB) sebelum alokasi IP baru.
- **VPN Client delete: peer tidak dihapus dari `wg.conf` di VPS** ([`db9ae7a`]) — Handler DELETE untuk `vpnServerId === '__vps_wg_server__'` hanya menghapus record DB tanpa menghapus `[Peer]` di `wg0.conf`. Fix: tambah call ke `POST /api/network/vps-wg-peer` dengan `action: 'remove'` sebelum delete DB.
- **Auto-create NAS/router saat tambah VPN client WireGuard** ([`701bfb7`]) — Endpoint `vps-wg-peer` secara otomatis membuat NAS record dan router saat tambah peer. Fix: hapus blok auto-create — NAS dikelola terpisah.
- **Auto-create NAS/router saat tambah VPN client L2TP** ([`8303308`]) — Sama seperti WireGuard, endpoint `vps-l2tp-peer` juga membuat NAS otomatis. Fix: hapus blok auto-create.
- **Panel redundansi di halaman VPN Client & VPN Server masih tampil** ([`8303308`], [`096d446`]) — Panel "Setup RADIUS Redundancy" yang sudah diputuskan untuk dihapus masih ter-render karena ada sisa JSX dan komponen stub. Fix: komponen `VpnServerRedundancyPanel` dijadikan stub `return null`, semua JSX orphan dibersihkan.

### Changed
- **`update.sh`: safe zero-downtime update** ([`08eba82`], [`8ce6421`], sesi ini) — Perbaikan menyeluruh pada script update:
  - `.env` di-backup ke `/tmp/salfanet-env-backup-<timestamp>` sebelum `git reset --hard` (extra safety meski `.env` ada di `.gitignore`)
  - Jika `.env` hilang setelah git reset, otomatis restore dari backup terakhir
  - Cleanup direktori orphan dari deployment lama (`srcappadmin`, `srclocales`, dll.) otomatis tiap update
  - PM2 `reload` (rolling zero-downtime) tetap digunakan saat restart — sesi PPPoE/Hotspot aktif tidak terputus oleh update kode
  - PM2 direstart (safety net) bahkan jika build gagal — server tidak dibiarkan mati
  - Tmp env backup lama (>7 hari) dibersihkan otomatis
  - Komentar safety guarantee ditambahkan di header script

---

<!-- AUTO-CHANGELOG:END -->

See full changelog: [docs/getting-started/CHANGELOG.md](docs/getting-started/CHANGELOG.md)

## 📚 Documentation

| File | Description |
|------|-------------|
| [docs/INSTALLATION-GUIDE.md](docs/INSTALLATION-GUIDE.md) | Complete VPS installation |
| [docs/GENIEACS-GUIDE.md](docs/GENIEACS-GUIDE.md) | GenieACS TR-069 setup & WiFi management |
| [docs/AGENT_DEPOSIT_SYSTEM.md](docs/AGENT_DEPOSIT_SYSTEM.md) | Agent balance & deposit |
| [docs/RADIUS-CONNECTIVITY.md](docs/RADIUS-CONNECTIVITY.md) | RADIUS architecture |
| [docs/FREERADIUS-SETUP.md](docs/FREERADIUS-SETUP.md) | FreeRADIUS configuration guide |

## 📝 License

MIT License - Free for commercial and personal use

## 👨‍💻 Development

Built with ❤️ for Indonesian ISPs

**Important**: Always use `formatWIB()` and `toWIB()` functions when displaying dates to users.
