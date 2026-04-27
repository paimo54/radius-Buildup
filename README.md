<![CDATA[<div align="center">

# 🛜 SALFANET RADIUS

### Modern Billing & RADIUS Management System for ISP / RTRW.NET

[![Version](https://img.shields.io/badge/version-2.25.1-blue.svg)](#changelog)
[![Next.js](https://img.shields.io/badge/Next.js-16-black.svg)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](#license)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#development)

Full-stack billing & network management platform with **FreeRADIUS** integration,
supporting **PPPoE**, **Hotspot**, **VPN**, and **FTTH/OLT** — built for Indonesian ISPs.

[Getting Started](#installation) · [Features](#features) · [Documentation](#documentation) · [Changelog](#changelog)

</div>

---

## 📑 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Project Structure](#-project-structure)
- [Installation](#%EF%B8%8F-installation)
- [Environment Variables](#-environment-variables)
- [FreeRADIUS](#-freeradius)
- [Go Backend (API v2)](#-go-backend-api-v2)
- [Cron Jobs](#-cron-jobs)
- [Android APK Builder](#-android-apk-builder)
- [Security](#-security)
- [Common Commands](#%EF%B8%8F-common-commands)
- [Troubleshooting](#-troubleshooting)
- [Documentation](#-documentation)
- [Changelog](#-changelog)
- [License](#-license)

---

## 🎯 Features

| Category | Key Capabilities |
|----------|-----------------|
| **RADIUS / Auth** | FreeRADIUS 3.0.26, PAP/CHAP/MS-CHAP, VPN L2TP/IPSec, PPPoE & Hotspot, CoA real-time speed/disconnect |
| **VPN Management** | MikroTik CHR via API, VPS built-in WireGuard & L2TP/IPsec, configurable IP pool & gateway, auto-generated RouterOS scripts |
| **PPPoE Management** | Customer accounts, profile-based bandwidth, isolation, IP assignment, MikroTik auto-sync, foto KTP+instalasi via kamera HP, GPS otomatis |
| **Hotspot Voucher** | 8 code types, batch up to 25,000, agent distribution, auto-sync with RADIUS, print templates |
| **Billing** | Postpaid/prepaid invoices, auto-generation, payment reminders, balance/deposit, auto-renewal |
| **Payment** | Manual upload (bukti transfer), Midtrans/Xendit/Duitku/Tripay gateway, approval workflow |
| **Notifications** | WhatsApp (Fonnte/WAHA/GOWA/MPWA/Wablas/WABlast/Kirimi.id), Email SMTP, broadcast, webhook |
| **Agent/Reseller** | Balance-based voucher generation, commission tracking, sales stats |
| **Financial** | Income/expense tracking with categories, reconciliation |
| **Network (FTTH)** | OLT/ODC/ODP management, customer port assignment, network map, SNMP monitoring |
| **GenieACS TR-069** | CPE/ONT management, WiFi config (SSID/password), device status & uptime |
| **Isolation** | Auto-isolate expired customers, customizable WhatsApp/Email/HTML landing page templates |
| **Cron Jobs** | 16 automated background jobs (tsx runner via PM2), history, distributed locking, manual trigger |
| **Roles & Permissions** | 53 permissions, 5 portals (Admin/Customer/Agent/Technician + SuperAdmin) |
| **PWA** | Installable across all portals, offline fallback, service worker cache |
| **Web Push** | VAPID-based browser push notifications, subscribe/unsubscribe, admin broadcast |
| **Android APK** | Build APK (WebView wrapper) directly on VPS for 4 portals, no GitHub Actions needed |
| **Bahasa** | Bahasa Indonesia (full i18n) |

---

## 🚀 Tech Stack

### Frontend (Next.js)

| Component | Technology |
|-----------|-----------|
| Framework | **Next.js 16** (App Router, standalone output, React Compiler) |
| Language | TypeScript 5 |
| Styling | Tailwind CSS 4 |
| UI Components | Radix UI + shadcn/ui |
| State | Zustand |
| Charts | Recharts |
| Maps | Leaflet / OpenStreetMap |
| Validation | Zod |
| Auth | NextAuth.js v4 + JWT |

### Backend (Go — API v2)

| Component | Technology |
|-----------|-----------|
| Framework | **Gin** v1.12 |
| ORM | GORM (MySQL driver) |
| Auth | golang-jwt/jwt v5 |
| SNMP | gosnmp |
| MikroTik | go-routeros + telnet |
| Scheduler | robfig/cron v3 |

### Infrastructure

| Component | Technology |
|-----------|-----------|
| Database | MySQL 8.0 + **Prisma ORM** (90 models) |
| RADIUS | FreeRADIUS 3.0.26 |
| Process Manager | PM2 (cluster mode) |
| Web Server | Nginx (reverse proxy + SSL) |
| Session Tracking | FreeRADIUS radacct (real-time) |

---

## 🏗 Architecture

```
┌──────────────────────────────────────────────────────────┐
│                      CLIENTS                             │
│  Admin Panel · Customer Portal · Agent · Technician      │
│  PWA (Web) · Android APK (WebView)                       │
└──────────────┬───────────────────────────┬───────────────┘
               │ HTTPS (Nginx)            │
               ▼                          ▼
┌──────────────────────┐    ┌──────────────────────┐
│   Next.js 16 (:3000) │    │   Go/Gin API (:8080) │
│   App Router + API   │    │   SNMP · MikroTik    │
│   44 API namespaces  │    │   OLT Monitoring     │
│   Prisma ORM         │    │   Cron · Dashboard   │
└──────────┬───────────┘    └──────────┬───────────┘
           │                           │
           ▼                           ▼
┌──────────────────────────────────────────────────────────┐
│                    MySQL 8.0 (90 tables)                 │
│  RADIUS tables (radcheck/radreply/radacct/radusergroup)  │
│  Business tables (pppoe_users/invoices/payments/agents)  │
│  Network tables (nas/vpn_servers/olt_onus/odp/odc)       │
└──────────────────────────┬───────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────┐
│                  FreeRADIUS 3.0.26                       │
│  SQL module → MySQL · REST module → Next.js API          │
│  PPPoE auth · Hotspot voucher · CoA (port 3799)          │
└──────────────────────────┬───────────────────────────────┘
                           │ RADIUS (1812/1813 UDP)
                           ▼
┌──────────────────────────────────────────────────────────┐
│              MikroTik Routers / NAS                      │
│  PPPoE Server · Hotspot · VPN (L2TP/WireGuard/SSTP)     │
└──────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
radius-Buildup/
├── src/
│   ├── app/
│   │   ├── admin/            # Admin panel (28 modules)
│   │   ├── agent/            # Agent/reseller portal
│   │   ├── api/              # 44 API route namespaces
│   │   ├── customer/         # Customer self-service portal
│   │   ├── technician/       # Technician portal
│   │   ├── daftar/           # Public registration
│   │   ├── evoucher/         # E-voucher purchase
│   │   ├── isolated/         # Isolation landing page
│   │   └── pay/              # Payment pages
│   ├── server/               # DB, services, jobs, cache, auth
│   ├── features/             # Vertical slices (queries, schemas, types)
│   ├── components/           # Shared React components
│   ├── cron/                 # TypeScript cron runner (tsx)
│   ├── locales/              # i18n (Bahasa Indonesia)
│   └── types/                # Shared TypeScript types
├── backend/                  # Go API (Gin + GORM)
│   ├── cmd/server/           # Entry point
│   └── internal/
│       ├── config/           # Environment config
│       ├── database/         # MySQL connection
│       ├── middleware/       # CORS, JWT auth
│       ├── models/           # GORM models
│       └── modules/          # 17 domain modules
│           ├── pppoe/        # PPPoE management
│           ├── hotspot/      # Hotspot vouchers
│           ├── olt/          # OLT SNMP monitoring
│           ├── mikrotik/     # RouterOS API
│           ├── billing/      # Invoices & payments
│           ├── dashboard/    # Analytics
│           └── ...           # auth, cron, sessions, etc.
├── prisma/
│   ├── schema.prisma         # Database schema (90 models)
│   └── seeds/                # Seed scripts
├── freeradius-config/        # FreeRADIUS config templates
├── vps-install/              # One-command VPS installer (22 scripts)
├── production/               # PM2, Nginx, deployment configs
├── scripts/                  # Utility & maintenance scripts
├── public/                   # PWA manifests, service worker, assets
├── tests/                    # Vitest test suites
└── docs/                     # Documentation & AI memory
```

---

## ⚙️ Installation

### Prerequisites

- Ubuntu 20.04+ / Debian 11+ (VPS or VM)
- Minimum 2GB RAM (4GB recommended)
- MySQL 8.0
- Node.js 20+
- Go 1.26+ (for backend API v2)

### Method 1 — Git Clone (Recommended)

```bash
ssh root@YOUR_VPS_IP

git clone https://github.com/paimo54/radius-Buildup.git /root/salfanet-radius
cd /root/salfanet-radius
bash vps-install/vps-installer.sh
```

The installer runs **interactively** — auto-detects environment, guides configuration, then executes all steps (MySQL → Node.js → FreeRADIUS → Nginx → PM2 → Security).

### Method 2 — Manual Upload via SCP

```bash
# From LOCAL terminal
scp -r ./salfanet-radius root@YOUR_VPS_IP:/root/salfanet-radius

# SSH to server
ssh root@YOUR_VPS_IP
cd /root/salfanet-radius
bash vps-install/vps-installer.sh
```

### Supported Environments

| Environment | Flag | Access |
|------------|------|--------|
| **Public VPS** (DigitalOcean, Vultr, Hetzner) | `--env vps` | Internet |
| **Proxmox LXC** | `--env lxc` | LAN/VLAN |
| **Proxmox VM / VirtualBox** | `--env vm` | LAN |
| **Bare Metal** | `--env bare` | LAN |

```bash
# Example: force environment + IP
bash vps-install/vps-installer.sh --env lxc --ip 192.168.1.50
```

### Starting the Go Backend

```bash
cd /root/salfanet-radius/backend
go run ./cmd/server
```

### Default Credentials

| | |
|--|--|
| Admin URL | `http://YOUR_VPS_IP/admin/login` |
| Username | `superadmin` |
| Password | `admin123` |

> ⚠️ **Change the password immediately after first login!**

### Updating

```bash
bash /var/www/salfanet-radius/vps-install/updater.sh
```

All uploaded data (logos, customer photos, payment receipts) and `.env` are automatically preserved during updates.

---

## 🔑 Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Database
DATABASE_URL="mysql://username:password@localhost:3306/database_name?connection_limit=10&pool_timeout=20"

# Timezone (WIB/WITA/WIT)
TZ="Asia/Jakarta"
NEXT_PUBLIC_TIMEZONE="Asia/Jakarta"

# App
NEXT_PUBLIC_APP_NAME="Your ISP Name"
NEXT_PUBLIC_APP_URL="http://localhost:3000"

# Auth secrets (generate with: openssl rand -base64 32)
NEXTAUTH_SECRET=<your-secret>
NEXTAUTH_URL=http://localhost:3000
AGENT_JWT_SECRET=<your-agent-secret>

# Encryption (generate with: openssl rand -hex 16)
ENCRYPTION_KEY=<your-32-char-hex-key>

# Web Push — generate with: npm run push:vapid
VAPID_PUBLIC_KEY="..."
VAPID_PRIVATE_KEY="..."
VAPID_CONTACT_EMAIL="admin@example.com"
```

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

### Auth Flow

**PPPoE:** `MikroTik → FreeRADIUS → MySQL (radcheck/radusergroup/radgroupreply)` → Access-Accept with Mikrotik-Rate-Limit

**Hotspot:** Same RADIUS path + `REST /api/radius/post-auth` → sets firstLoginAt, expiresAt

### RADIUS Tables

| Table | Purpose |
|-------|---------|
| `radcheck` | User credentials |
| `radreply` | User-specific reply attributes |
| `radusergroup` | User → Group mapping |
| `radgroupreply` | Group reply (bandwidth, session timeout) |
| `radacct` | Session accounting |
| `nas` | NAS/Router clients (dynamic) |

### CoA (Change of Authorization)

Sends real-time speed/disconnect commands to MikroTik without dropping PPPoE connections.

**MikroTik:** `/radius incoming set accept=yes port=3799`

**API:** `POST /api/radius/coa` — actions: `disconnect`, `update`, `sync-profile`, `test`

---

## 🔧 Go Backend (API v2)

The Go backend runs alongside Next.js, handling performance-critical modules:

```
backend/internal/modules/
├── activity/     # Activity logging
├── admin/        # Admin user management
├── auth/         # JWT authentication
├── billing/      # Invoice & payment processing
├── company/      # Company settings
├── cron/         # Background job scheduler
├── customer/     # Customer portal API
├── dashboard/    # Analytics & statistics
├── health/       # Health check endpoint
├── hotspot/      # Hotspot voucher management
├── mikrotik/     # RouterOS API integration
├── olt/          # OLT SNMP monitoring & ONU sync
├── payment/      # Payment gateway callbacks
├── pppoe/        # PPPoE user management
├── router/       # NAS/Router CRUD
├── sessions/     # Active session tracking
└── whatsapp/     # WhatsApp notification dispatch
```

**Key dependencies:** Gin (HTTP), GORM (ORM), gosnmp (SNMP), go-routeros (MikroTik API), robfig/cron (scheduler)

---

## ⏰ Cron Jobs

16 automated jobs managed by the TypeScript cron runner (`src/cron/runner.ts`):

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

## 📱 Android APK Builder

Build Android APKs (WebView wrapper) for all 4 portals directly on your VPS — no Android Studio required.

### Setup Android SDK (one-time)

```bash
apt-get update && apt-get install -y openjdk-17-jdk wget unzip && \
mkdir -p /opt/android/cmdline-tools && \
wget -q https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip -O /tmp/cmdtools.zip && \
unzip -q /tmp/cmdtools.zip -d /opt/android/cmdline-tools && \
mv /opt/android/cmdline-tools/cmdline-tools /opt/android/cmdline-tools/latest && \
yes | /opt/android/cmdline-tools/latest/bin/sdkmanager --licenses && \
/opt/android/cmdline-tools/latest/bin/sdkmanager "platforms;android-34" "build-tools;34.0.0" && \
echo 'export ANDROID_HOME=/opt/android' >> /etc/environment
```

### Build & Download

Navigate to **Admin → Download Aplikasi Android** → click **Build APK**.

| Role | Package ID | Theme Color |
|------|-----------|-------------|
| Admin | `net.salfanet.admin` | Blue |
| Customer | `net.salfanet.customer` | Cyan |
| Technician | `net.salfanet.technician` | Green |
| Agent | `net.salfanet.agent` | Purple |

---

## 🔐 Security

The installer automatically configures:

- **fail2ban** — Brute-force protection (SSH, Nginx)
- **UFW Firewall** — Default deny, allow only required ports
- **Disk cleanup cron** — Daily log rotation & temp cleanup

### Required Firewall Ports

```bash
ufw allow 22/tcp     # SSH
ufw allow 80/tcp     # HTTP
ufw allow 443/tcp    # HTTPS
ufw allow 1812/udp   # RADIUS Auth
ufw allow 1813/udp   # RADIUS Accounting
ufw allow 3799/udp   # RADIUS CoA
```

### Post-Install Checklist

1. ✅ Change default admin password
2. ✅ Update MySQL passwords in `.env`
3. ✅ Configure SSL (Let's Encrypt or Cloudflare)
4. ✅ Generate unique `NEXTAUTH_SECRET` and `AGENT_JWT_SECRET`
5. ✅ Set `ENCRYPTION_KEY` for sensitive data at rest

---

## 🛠️ Common Commands

```bash
# PM2
pm2 status
pm2 logs salfanet-radius --lines 100
pm2 restart ecosystem.config.js --update-env

# FreeRADIUS
systemctl restart freeradius
freeradius -XC                                    # Test config
radtest 'user@realm' password 127.0.0.1 0 testing123

# Database
mysqldump -u salfanet_user -p salfanet_radius > backup.sql

# Go Backend
cd backend && go run ./cmd/server

# Development
npm run dev          # Next.js dev server
npm run test         # Run Vitest
npm run db:push      # Sync Prisma schema to DB
npm run db:seed      # Seed initial data
npm run cron         # Run cron service standalone
```

---

## 🧯 Troubleshooting

### Website not accessible from VPS IP

```bash
# Check services on VPS
ss -tulpn | grep -E ':80|:443|:3000'
curl -I http://127.0.0.1:3000
systemctl status nginx --no-pager
pm2 status
```

If all local checks pass, verify port forwarding: `Public:80 → VM:80`, `Public:443 → VM:443`.

### PM2 running but blank page

```bash
cd /var/www/salfanet-radius
npm run build
pm2 restart ecosystem.config.js --update-env
```

### Run Nginx diagnostics

```bash
bash vps-install/install-nginx.sh
```

---

## 📲 WhatsApp Providers

| Provider | Base URL | Auth |
|----------|----------|------|
| Fonnte | `https://api.fonnte.com/send` | Token |
| WAHA | `http://IP:PORT` | API Key |
| GOWA | `http://IP:PORT` | `user:pass` |
| MPWA | `http://IP:PORT` | API Key |
| Wablas | `https://pati.wablas.com` | Token |
| Kirimi.id | `https://api.kirimi.id` | API Key |

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

Dashboard · PPPoE · Hotspot · Agent · Invoice · Payment · Keuangan · Sessions · WhatsApp · Network (OLT/ODC/ODP) · GenieACS · Inventory · Notifications · Push Notifications · Tickets · Settings · System

**Roles:** `SUPER_ADMIN` · `FINANCE` · `CUSTOMER_SERVICE` · `TECHNICIAN` · `MARKETING` · `VIEWER`

---

## 📝 Changelog

See full changelog: [CHANGELOG.md](CHANGELOG.md)

<!-- AUTO-CHANGELOG:START -->

### v2.25.1 — 2026-04-26

**Added:**
- `install-security.sh` — Automated server security module (fail2ban + UFW + disk cleanup cron)

**Fixed:**
- Disk full 100% causing MySQL deadlock & API 500
- APK build connection pool exhaustion during concurrent builds

---

### v2.25.0 — 2026-04-26

**Added:**
- Build APK Android directly on VPS server (no GitHub Actions / Android Studio required)
- Admin panel UI for build management with auto-polling status

---

### v2.24.0 — 2026-04-26

**Removed:**
- Web-based update panel (replaced with SSH-based `updater.sh`)

---

### v2.23.0 — 2026-04-26

**Added:**
- TypeScript cron runner (`src/cron/runner.ts`) replacing `cron-service.js`
- PM2 ecosystem config template

**Removed:**
- Coordinator role (never fully implemented)
- Firebase Admin SDK & FCM (replaced with VAPID Web Push)

<!-- AUTO-CHANGELOG:END -->

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [docs/AI_PROJECT_MEMORY.md](docs/AI_PROJECT_MEMORY.md) | Full architecture, VPS details, DB schema, known issues |
| [docs/COMPREHENSIVE_FEATURE_GUIDE.md](docs/COMPREHENSIVE_FEATURE_GUIDE.md) | Complete feature documentation |
| [docs/DOCS_INDEX.md](docs/DOCS_INDEX.md) | Documentation index |
| [docs/ISOLATION_SYSTEM.md](docs/ISOLATION_SYSTEM.md) | PPPoE isolation system guide |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Development roadmap |
| [docs/SECURITY_FIXES_APPLIED.md](docs/SECURITY_FIXES_APPLIED.md) | Security audit & fixes |
| [freeradius-config/README.md](freeradius-config/README.md) | FreeRADIUS configuration guide |
| [vps-install/README.md](vps-install/README.md) | VPS installer documentation |
| [production/PRODUCTION_DEPLOYMENT.md](production/PRODUCTION_DEPLOYMENT.md) | Production deployment guide |

---

## 🤖 AI Development Assistant

If you're an AI assistant working on this project, **read first:**

📖 [docs/AI_PROJECT_MEMORY.md](docs/AI_PROJECT_MEMORY.md) — contains full architecture, VPS details, DB schema, known issues, and proven solutions.

**Key rules:**
- Always use `formatWIB()` and `toWIB()` when displaying dates
- Database stores UTC, frontend displays WIB (UTC+7)
- Prisma schema has 90 models — check before adding new ones
- FreeRADIUS tables (`radcheck`, `radreply`, etc.) follow strict naming conventions

---

## 📝 License

MIT License — Free for commercial and personal use.

---

<div align="center">

Built with ❤️ for Indonesian ISPs

**[⬆ Back to Top](#-salfanet-radius)**

</div>
]]>
