#!/bin/bash
# ============================================================================
# SALFANET RADIUS - VPS Updater
# ============================================================================
# Update existing installation to the latest GitHub release.
#
# Usage:
#   bash updater.sh                         # Update to latest release
#   bash updater.sh --version v2.12.0       # Update to specific version
#   bash updater.sh --branch master         # Update from git branch (no build)
#   bash updater.sh --skip-backup           # Skip pre-update backup
# ============================================================================

set -e
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ─── Colors ────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; WHITE='\033[1;37m'; NC='\033[0m'

print_step()    { echo -e "\n${CYAN}▶ $1${NC}"; }
print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_info()    { echo -e "${YELLOW}  $1${NC}"; }
print_error()   { echo -e "${RED}✗ $1${NC}" >&2; }

# ─── Config ────────────────────────────────────────────────────────────────
APP_DIR="${APP_DIR:-/var/www/salfanet-radius}"
GITHUB_REPO="s4lfanet/salfanet-radius"
PM2_APP_NAME="salfanet-radius"
PM2_CRON_NAME="salfanet-cron"
BACKUP_BASE="/root/salfanet-backups"
TARGET_VERSION=""
USE_BRANCH=""
SKIP_BACKUP=false
ARCH="amd64"

# Detect architecture
if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
    ARCH="arm64"
fi

# ─── Parse args ────────────────────────────────────────────────────────────
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --version)     TARGET_VERSION="$2"; shift ;;
        --branch)      USE_BRANCH="$2"; shift ;;
        --skip-backup) SKIP_BACKUP=true ;;
        --app-dir)     APP_DIR="$2"; shift ;;
        --help|-h)
            echo "Usage: bash updater.sh [--version vX.Y.Z] [--branch master] [--skip-backup]"
            exit 0 ;;
    esac
    shift
done

# Default to git branch mode (no GitHub Releases used)
if [ -z "$USE_BRANCH" ] && [ -z "$TARGET_VERSION" ]; then
    USE_BRANCH="master"
fi

# ─── Sanity checks ─────────────────────────────────────────────────────────
if [ "$EUID" -ne 0 ]; then
    print_error "Run as root: sudo bash updater.sh"
    exit 1
fi

if [ ! -d "$APP_DIR" ]; then
    print_error "App not found at $APP_DIR. Run the installer first."
    exit 1
fi

# ─── Show current version ──────────────────────────────────────────────────
CURRENT_VERSION="unknown"
if [ -f "$APP_DIR/VERSION" ]; then
    CURRENT_VERSION=$(cat "$APP_DIR/VERSION")
elif [ -f "$APP_DIR/package.json" ]; then
    CURRENT_VERSION=$(node -p "require('$APP_DIR/package.json').version" 2>/dev/null || echo "unknown")
fi

echo ""
echo -e "${WHITE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║      SALFANET RADIUS — VPS Updater               ║${NC}"
echo -e "${WHITE}╚══════════════════════════════════════════════════╝${NC}"
echo ""
print_info "App dir      : $APP_DIR"
print_info "Current ver  : $CURRENT_VERSION"
print_info "Architecture : $ARCH"
echo ""

# ──────────────────────────────────────────────────────────────────────────
# MODE A: Update via git pull (branch mode, no build download)
# ──────────────────────────────────────────────────────────────────────────
if [ -n "$USE_BRANCH" ]; then
    print_step "Updating via git branch: $USE_BRANCH"

    if [ ! -d "$APP_DIR/.git" ]; then
        print_error "Not a git repo at $APP_DIR. Use release mode instead."
        exit 1
    fi

    cd "$APP_DIR"

    # Backup
    if [ "$SKIP_BACKUP" = false ]; then
        print_step "Creating backup"
        BACKUP_DIR="$BACKUP_BASE/$(date +%Y%m%d-%H%M%S)-git"
        mkdir -p "$BACKUP_DIR"
        cp -r "$APP_DIR" "$BACKUP_DIR/app" 2>/dev/null || true
        print_success "Backup saved to $BACKUP_DIR"
    fi

    # ─── Migrate uploads to persistent directory ───────────────────────
    # Uploads now live in /var/data/salfanet/uploads/ (outside git/build).
    # Migrate any remaining files from legacy public/uploads/ location.
    UPLOAD_DIR="${UPLOAD_DIR:-/var/data/salfanet/uploads}"
    mkdir -p "$UPLOAD_DIR"

    # Add UPLOAD_DIR to .env if not present
    if [ -f "$APP_DIR/.env" ] && ! grep -q '^UPLOAD_DIR=' "$APP_DIR/.env"; then
        echo "" >> "$APP_DIR/.env"
        echo "# Persistent upload directory (survives rebuilds)" >> "$APP_DIR/.env"
        echo "UPLOAD_DIR=$UPLOAD_DIR" >> "$APP_DIR/.env"
        print_success "UPLOAD_DIR added to .env"
    fi

    # One-time migration from public/uploads/ to persistent dir
    if [ -d "$APP_DIR/public/uploads" ] && [ "$(ls -A "$APP_DIR/public/uploads" 2>/dev/null)" ]; then
        for subdir in "$APP_DIR/public/uploads"/*/; do
            [ -d "$subdir" ] || continue
            dirname=$(basename "$subdir")
            if [ "$(ls -A "$subdir" 2>/dev/null)" ]; then
                mkdir -p "$UPLOAD_DIR/$dirname"
                cp -rn "$subdir"* "$UPLOAD_DIR/$dirname/" 2>/dev/null || true
            fi
        done
        print_success "Uploads migrated to $UPLOAD_DIR (safe from rebuilds)"
    fi

    git fetch origin
    git reset --hard "origin/$USE_BRANCH"
    git clean -fd

    # ── Update ecosystem.config.js (untracked by git) ─────────────────────
    # ecosystem.config.js is untracked — git clean removes it, must be restored.
    ECOSYSTEM_CHANGED=false
    if [ -f "$APP_DIR/production/ecosystem.config.js" ]; then
        OLD_SCRIPT=$(grep -o "script:.*'[^']*'" "$APP_DIR/ecosystem.config.js" 2>/dev/null | grep -i cron | head -1 || echo "")
        cp "$APP_DIR/production/ecosystem.config.js" "$APP_DIR/ecosystem.config.js"
        NEW_SCRIPT=$(grep -o "script:.*'[^']*'" "$APP_DIR/ecosystem.config.js" 2>/dev/null | grep -i cron | head -1 || echo "")
        if [ "$OLD_SCRIPT" != "$NEW_SCRIPT" ]; then
            ECOSYSTEM_CHANGED=true
        fi
        print_success "ecosystem.config.js updated from production/"
    fi

    # ── Cleanup stale files from refactor phases ────────────────────────
    print_step "Cleaning up stale files from refactor"
    for stale in \
        "src/server/push.service.ts" \
        "src/server/push.service.js" \
        "firebase-service-account.json" \
        "src/lib/cron" \
        "src/app/coordinator" \
        "src/app/admin/coordinators" \
        "src/app/api/billing" \
        "src/app/api/cron/history" \
        "src/app/api/settings/telegram-backup" \
        "src/components/dashboard" \
        "chk-pg.js" "deploy.sh" "bad-files.txt" "start-dev.ps1" "kill-ports.ps1"; do
        if [ -e "$APP_DIR/$stale" ]; then
            rm -rf "${APP_DIR:?}/$stale"
            print_info "Removed stale: $stale"
        fi
    done
    print_success "Stale file cleanup done"

    print_step "Installing dependencies"
    # Try npm ci first (faster, strict lock file) — fall back to npm install
    # if lock file is out of sync with package.json (common after refactor).
    if ! npm ci --omit=dev 2>/tmp/updater-npm-ci.log; then
        print_info "npm ci failed (lock file mismatch) — falling back to npm install..."
        npm install --production=false 2>&1 | tail -10
    fi

    print_step "Generating Prisma client"
    node_modules/.bin/prisma generate

    print_step "Running database migrations"
    node_modules/.bin/prisma db push --accept-data-loss 2>/dev/null || node_modules/.bin/prisma db push

    print_step "Building application"
    NODE_OPTIONS="--max-old-space-size=1536" NEXT_TELEMETRY_DISABLED=1 npm run build

    # ── Copy static assets to standalone ──────────────────────────────────
    if [ -d "$APP_DIR/.next/standalone" ]; then
        mkdir -p "$APP_DIR/.next/standalone/public"
        cp -r "$APP_DIR/public/." "$APP_DIR/.next/standalone/public/" 2>/dev/null || true
        mkdir -p "$APP_DIR/.next/standalone/.next"
        cp -r "$APP_DIR/.next/static" "$APP_DIR/.next/standalone/.next/static/" 2>/dev/null || true
        print_success "Static assets copied to standalone"
    fi

    print_step "Restarting services"
    pm2 reload "$PM2_APP_NAME" --update-env 2>/dev/null || pm2 restart "$PM2_APP_NAME" 2>/dev/null || true

    # Jika ecosystem.config.js berubah (migrasi cron-service.js → tsx runner),
    # harus delete + start ulang agar PM2 pakai script/args baru.
    if [ "${ECOSYSTEM_CHANGED:-false}" = true ]; then
        print_info "Ecosystem config changed — migrating salfanet-cron ke tsx runner..."
        pm2 delete "$PM2_CRON_NAME" 2>/dev/null || true
        pm2 start "$APP_DIR/ecosystem.config.js" --only "$PM2_CRON_NAME" 2>&1 | tail -3
        print_success "salfanet-cron migrated to tsx runner"
    else
        pm2 restart "$PM2_CRON_NAME" --update-env 2>/dev/null || true
    fi
    pm2 save

    # ─── Security: pastikan fail2ban + UFW + cleanup cron terpasang ──────
    if [ -f "$APP_DIR/vps-install/install-security.sh" ]; then
        source "$APP_DIR/vps-install/install-security.sh"
        # Hanya setup cleanup cron dan fail2ban (UFW sudah dikonfigurasi saat install)
        setup_cleanup_cron 2>/dev/null || true
        # Pastikan fail2ban running jika sudah terinstall
        if command -v fail2ban-client &>/dev/null; then
            systemctl is-active --quiet fail2ban || systemctl restart fail2ban 2>/dev/null || true
            print_success "fail2ban status: $(systemctl is-active fail2ban 2>/dev/null)"
        fi
    fi

    # ─── VPN post-update (sama seperti mode release) ──────────────────────
    # VPN Client (CHR forwarding)
    if [ -f "/usr/local/bin/vpn-connect" ]; then
        REINSTALL_VPN=true bash "$APP_DIR/vps-install/install-vpn-client.sh" 2>/dev/null \
            && print_success "VPN client (CHR mode) helper diperbarui" || true
    fi
    # WireGuard Server
    if [ -f "/etc/wireguard/wg-server-info.json" ] && [ -f "$APP_DIR/vps-install/install-wg-server.sh" ]; then
        WG_IFACE=$(grep -o '"interface": *"[^"]*"' /etc/wireguard/wg-server-info.json 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "wg0")
        WG_PORT=$(grep -o '"listenPort": *[0-9]*' /etc/wireguard/wg-server-info.json 2>/dev/null | grep -o '[0-9]*$' || echo "51820")
        WG_SUBNET=$(grep -o '"subnet": *"[^"]*"' /etc/wireguard/wg-server-info.json 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "10.200.0.0/24")
        WG_IFACE="$WG_IFACE" WG_PORT="$WG_PORT" WG_SUBNET="$WG_SUBNET" \
            bash "$APP_DIR/vps-install/install-wg-server.sh" 2>/dev/null \
            && print_success "WireGuard server diperbarui" || true
    fi
    # L2TP/IPsec Server
    if [ -f "/etc/salfanet/l2tp/l2tp-server-info.json" ] && [ -f "$APP_DIR/vps-install/install-l2tp-server.sh" ]; then
        L2TP_PSK=$(grep -o '"ipsecPsk": *"[^"]*"' /etc/salfanet/l2tp/l2tp-server-info.json 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "")
        L2TP_SUBNET=$(grep -o '"subnet": *"[^"]*"' /etc/salfanet/l2tp/l2tp-server-info.json 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "10.201.0.0/24")
        L2TP_PSK="$L2TP_PSK" L2TP_SUBNET="$L2TP_SUBNET" \
            bash "$APP_DIR/vps-install/install-l2tp-server.sh" 2>/dev/null \
            && print_success "L2TP/IPsec server diperbarui" || true
    fi

    NEW_VERSION=$(node -p "require('$APP_DIR/package.json').version" 2>/dev/null || echo "unknown")
    echo ""
    print_success "Update complete! ${CURRENT_VERSION} → ${NEW_VERSION}"
    exit 0
fi

# ──────────────────────────────────────────────────────────────────────────
# MODE B: Update via GitHub Release webupload ZIP
# ──────────────────────────────────────────────────────────────────────────
print_step "Fetching release information"

if [ -z "$TARGET_VERSION" ]; then
    # Get latest release tag from GitHub API
    LATEST=$(curl -sSf "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
        | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
    if [ -z "$LATEST" ]; then
        print_error "Could not fetch latest release from GitHub. Check internet connectivity."
        exit 1
    fi
    TARGET_VERSION="$LATEST"
fi

print_info "Target version : $TARGET_VERSION"

# Check if already on target version
if [ "$CURRENT_VERSION" = "$TARGET_VERSION" ] || [ "v$CURRENT_VERSION" = "$TARGET_VERSION" ]; then
    echo ""
    print_success "Already on version $TARGET_VERSION — nothing to update."
    echo "Use --version to force a specific version or --branch to update from git."
    exit 0
fi

# ─── Download webupload ZIP ────────────────────────────────────────────────
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${TARGET_VERSION}/webupload-${ARCH}.zip"
WORK_DIR=$(mktemp -d)
ZIP_PATH="$WORK_DIR/webupload.zip"

print_step "Downloading webupload-${ARCH}.zip (${TARGET_VERSION})"
print_info "URL: $DOWNLOAD_URL"

if ! curl -sSfL --progress-bar -o "$ZIP_PATH" "$DOWNLOAD_URL"; then
    print_error "Download failed. Check the version tag and network."
    rm -rf "$WORK_DIR"
    exit 1
fi

print_success "Downloaded $(du -sh "$ZIP_PATH" | cut -f1)"

# ─── Backup current app ────────────────────────────────────────────────────
if [ "$SKIP_BACKUP" = false ]; then
    print_step "Backing up current installation"
    BACKUP_DIR="$BACKUP_BASE/$(date +%Y%m%d-%H%M%S)-${CURRENT_VERSION}"
    mkdir -p "$BACKUP_DIR"

    # Only backup app code (skip uploads/ and node_modules/ — too large)
    rsync -a --exclude='node_modules' --exclude='uploads' \
        "$APP_DIR/" "$BACKUP_DIR/app/" 2>/dev/null || \
        cp -r "$APP_DIR" "$BACKUP_DIR/app"

    print_success "Backup saved to $BACKUP_DIR"
fi

# ─── Extract & stage ───────────────────────────────────────────────────────
print_step "Extracting new build"
EXTRACT_DIR="$WORK_DIR/extracted"
mkdir -p "$EXTRACT_DIR"
unzip -q "$ZIP_PATH" -d "$EXTRACT_DIR"

# The zip contains webupload-staging/ as root dir
STAGED_DIR=$(find "$EXTRACT_DIR" -maxdepth 1 -type d | grep -v "^$EXTRACT_DIR$" | head -1)
if [ -z "$STAGED_DIR" ]; then
    print_error "Unexpected zip structure."
    rm -rf "$WORK_DIR"
    exit 1
fi

# ─── Ensure required system packages ─────────────────────────────────────
print_step "Checking system dependencies"
MISSING_PKGS=""
for pkg in sshpass xl2tpd; do
    if ! dpkg -s "$pkg" &>/dev/null; then
        MISSING_PKGS="$MISSING_PKGS $pkg"
    fi
done
# Jika WireGuard server sudah terinstall, pastikan paket wg tersedia
if [ -f "/etc/wireguard/wg-server-info.json" ]; then
    for pkg in wireguard wireguard-tools; do
        if ! dpkg -s "$pkg" &>/dev/null; then
            MISSING_PKGS="$MISSING_PKGS $pkg"
        fi
    done
fi
# Jika L2TP server sudah terinstall, pastikan strongswan+xl2tpd tersedia
if [ -f "/etc/salfanet/l2tp/l2tp-server-info.json" ]; then
    for pkg in strongswan xl2tpd; do
        if ! dpkg -s "$pkg" &>/dev/null; then
            MISSING_PKGS="$MISSING_PKGS $pkg"
        fi
    done
fi
if [ -n "$MISSING_PKGS" ]; then
    print_info "Installing missing packages:$MISSING_PKGS"
    apt-get install -y $MISSING_PKGS || print_info "Warning: some packages could not be installed"
else
    print_success "System packages OK (sshpass, xl2tpd, wg tools if applicable)"
fi

# ─── Stop services ────────────────────────────────────────────────────────
print_step "Stopping PM2 processes"
pm2 stop "$PM2_APP_NAME" 2>/dev/null || true
pm2 stop "$PM2_CRON_NAME" 2>/dev/null || true
print_success "Services stopped"

# ─── Deploy new build ─────────────────────────────────────────────────────
print_step "Deploying new build"

# Preserve .env from current installation
ENV_FILE=""
if [ -f "$APP_DIR/.env" ]; then
    ENV_FILE=$(mktemp)
    cp "$APP_DIR/.env" "$ENV_FILE"
fi

# Migrate uploads to persistent directory before replacing files
UPLOAD_DIR="${UPLOAD_DIR:-/var/data/salfanet/uploads}"
mkdir -p "$UPLOAD_DIR"
if [ -d "$APP_DIR/public/uploads" ] && [ "$(ls -A "$APP_DIR/public/uploads" 2>/dev/null)" ]; then
    for subdir in "$APP_DIR/public/uploads"/*/; do
        [ -d "$subdir" ] || continue
        dirname=$(basename "$subdir")
        if [ "$(ls -A "$subdir" 2>/dev/null)" ]; then
            mkdir -p "$UPLOAD_DIR/$dirname"
            cp -rn "$subdir"* "$UPLOAD_DIR/$dirname/" 2>/dev/null || true
        fi
    done
    print_success "Uploads migrated to $UPLOAD_DIR"
fi

# Replace app files (keep a few runtime dirs)
rsync -a --delete \
    --exclude='.env' \
    "$STAGED_DIR/" "$APP_DIR/"

# Restore .env
if [ -n "$ENV_FILE" ] && [ -f "$ENV_FILE" ]; then
    cp "$ENV_FILE" "$APP_DIR/.env"
    rm -f "$ENV_FILE"
    print_success ".env restored"
fi

# Ensure UPLOAD_DIR in .env
if [ -f "$APP_DIR/.env" ] && ! grep -q '^UPLOAD_DIR=' "$APP_DIR/.env"; then
    echo "" >> "$APP_DIR/.env"
    echo "# Persistent upload directory (survives rebuilds)" >> "$APP_DIR/.env"
    echo "UPLOAD_DIR=$UPLOAD_DIR" >> "$APP_DIR/.env"
    print_success "UPLOAD_DIR added to .env"
fi

# ─── Run DB migrations ────────────────────────────────────────────────────
print_step "Running database migrations (prisma db push)"
cd "$APP_DIR"

if [ -f "$APP_DIR/.env" ]; then
    export $(grep -v '^#' "$APP_DIR/.env" | grep 'DATABASE_URL' | xargs) 2>/dev/null || true
fi

node_modules/.bin/prisma generate 2>/dev/null || true
node_modules/.bin/prisma db push --accept-data-loss 2>/dev/null || \
    node_modules/.bin/prisma db push || \
    print_info "DB push skipped (check manually)"

print_step "Applying seed data (new templates & config)"
npm run db:seed 2>/dev/null || print_info "Seed skipped (check manually)"

# ─── Update VPN Client (SSTP/L2TP ke CHR) jika sudah terinstall ──────────
# Flow lama: VPS sebagai client → konek ke MikroTik CHR → FreeRADIUS
# Tetap dipertahankan untuk deployment VPS lokal lewat CHR
VPN_CLIENT_CONF="/etc/vpn/vpn.conf"
if [ -f "$VPN_CLIENT_CONF" ] || systemctl is-active --quiet vpn-tunnel 2>/dev/null || [ -f "/usr/local/bin/vpn-connect" ]; then
    print_step "Update VPN Client (CHR forwarding — SSTP/L2TP client)"
    if [ -f "$APP_DIR/vps-install/install-vpn-client.sh" ]; then
        # Re-install hanya update helper scripts + service file, tidak reset konfigurasi
        REINSTALL_VPN=true bash "$APP_DIR/vps-install/install-vpn-client.sh" 2>/dev/null \
            && print_success "VPN client (CHR mode) helper diperbarui" \
            || print_info "VPN client update skipped (tidak kritis)"
    fi
fi

# ─── Update WireGuard Server jika sudah terinstall ───────────────────────
# Flow baru: VPS sebagai WireGuard server, NAS konek langsung
WG_INFO="/etc/wireguard/wg-server-info.json"
if [ -f "$WG_INFO" ]; then
    print_step "Update WireGuard VPN Server"
    WG_IFACE=$(grep -o '"interface": *"[^"]*"' "$WG_INFO" 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "wg0")
    WG_SUBNET=$(grep -o '"subnet": *"[^"]*"' "$WG_INFO" 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "10.200.0.0/24")
    WG_PORT=$(grep -o '"listenPort": *[0-9]*' "$WG_INFO" 2>/dev/null | grep -o '[0-9]*$' || echo "51820")

    if [ -f "$APP_DIR/vps-install/install-wg-server.sh" ]; then
        print_info "Re-running WireGuard server installer (idempotent, peers tidak terputus)"
        # jalankan dengan subnet/port yang sama dari info file yang ada
        WG_IFACE="$WG_IFACE" WG_PORT="$WG_PORT" WG_SUBNET="$WG_SUBNET" \
            bash "$APP_DIR/vps-install/install-wg-server.sh" \
            && print_success "WireGuard server diperbarui (wg-server-info.json + wg syncconf)" \
            || print_info "WireGuard update gagal — cek: systemctl status wg-quick@${WG_IFACE}"
    else
        # Jika tidak ada script (versi lama), pastikan service jalan saja
        systemctl is-active --quiet "wg-quick@${WG_IFACE}" \
            && print_success "WireGuard service masih aktif" \
            || { systemctl start "wg-quick@${WG_IFACE}" 2>/dev/null && print_success "WireGuard service direstart"; }
    fi
fi

# ─── Update L2TP/IPsec Server jika sudah terinstall ──────────────────────
# Fallback untuk RouterOS 6 yang tidak support WireGuard
L2TP_INFO="/etc/salfanet/l2tp/l2tp-server-info.json"
if [ -f "$L2TP_INFO" ]; then
    print_step "Update L2TP/IPsec VPN Server"
    if [ -f "$APP_DIR/vps-install/install-l2tp-server.sh" ]; then
        # Preserve PSK agar NAS tidak perlu rekonfigurasi
        L2TP_PSK=$(grep -o '"ipsecPsk": *"[^"]*"' "$L2TP_INFO" 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "")
        L2TP_SUBNET=$(grep -o '"subnet": *"[^"]*"' "$L2TP_INFO" 2>/dev/null | sed 's/.*: *"//;s/"//' || echo "10.201.0.0/24")

        print_info "Re-running L2TP server installer (PSK dipertahankan)"
        L2TP_PSK="$L2TP_PSK" L2TP_SUBNET="$L2TP_SUBNET" \
            bash "$APP_DIR/vps-install/install-l2tp-server.sh" \
            && print_success "L2TP/IPsec server diperbarui" \
            || print_info "L2TP update gagal — cek: systemctl status xl2tpd"
    else
        # Pastikan service jalan
        systemctl is-active --quiet xl2tpd \
            && print_success "xl2tpd service masih aktif" \
            || { systemctl start xl2tpd 2>/dev/null && print_success "xl2tpd direstart"; }
    fi
fi

# ─── Restart services ─────────────────────────────────────────────────────
print_step "Starting PM2 processes"
pm2 start "$PM2_APP_NAME" 2>/dev/null || true
pm2 start "$PM2_CRON_NAME" 2>/dev/null || true
pm2 save

# ─── Cleanup ──────────────────────────────────────────────────────────────
rm -rf "$WORK_DIR"

# ─── Done ─────────────────────────────────────────────────────────────────
NEW_VERSION=$(cat "$APP_DIR/VERSION" 2>/dev/null || echo "$TARGET_VERSION")
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        Update berhasil!                          ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
print_success "${CURRENT_VERSION}  →  ${NEW_VERSION}"
echo ""
print_info "Cek status   : pm2 status"
print_info "Cek log      : pm2 logs ${PM2_APP_NAME}"
print_info "Backup ada di: $BACKUP_BASE"
# Tampilkan status VPN
[ -f "/usr/local/bin/vpn-connect" ]               && print_info "VPN Client (CHR) : vpn-connect status"
[ -f "/etc/wireguard/wg-server-info.json" ]        && print_info "WireGuard Server : wg show wg0"
[ -f "/etc/salfanet/l2tp/l2tp-server-info.json" ]  && print_info "L2TP/IPsec Server: systemctl status xl2tpd"
echo ""
