# L2TP VPN Client Setup untuk Proxmox VPS

## Masalah

L2TP client memerlukan `/dev/ppp` dan `/dev/net/tun` devices yang tidak bisa dibuat dari dalam Proxmox LXC container karena security restrictions.

## Solusi

### Opsi 1: Enable TUN/TAP di Proxmox Host (Recommended)

Admin Proxmox perlu enable TUN/TAP support untuk container Anda.

**Di Proxmox Host, jalankan:**

```bash
# Ganti CTID dengan ID container Anda (misalnya 100, 101, dst)
CTID=YOUR_CONTAINER_ID

# Enable TUN device
pct set $CTID -features nesting=1,keyctl=1
pct set $CTID -dev0 /dev/net/tun,mode=0666

# Enable PPP device
pct set $CTID -dev1 /dev/ppp,mode=0600

# Load PPP modules di host
modprobe ppp_generic
modprobe ppp_async
modprobe ppp_deflate  
modprobe ppp_mppe
modprobe l2tp_ppp

# Restart container
pct reboot $CTID
```

**Atau edit config file manually:**

```bash
# Edit /etc/pve/lxc/CTID.conf
nano /etc/pve/lxc/$CTID.conf

# Tambahkan baris berikut:
features: keyctl=1,nesting=1
lxc.cgroup2.devices.allow: c 10:200 rwm
lxc.cgroup2.devices.allow: c 108:0 rwm
lxc.mount.entry: /dev/net/tun dev/net/tun none bind,create=file 0 0
lxc.mount.entry: /dev/ppp dev/ppp none bind,create=file 0 0

# Restart container
pct reboot $CTID
```

### Opsi 2: Gunakan KVM VM (Full Virtualization)

Jika menggunakan LXC container, pertimbangkan untuk migrasi ke KVM VM yang memiliki akses penuh ke kernel dan devices.

### Opsi 3: Alternative VPN - WireGuard (More Container-Friendly)

WireGuard lebih compatible dengan containers dan tidak memerlukan `/dev/ppp`:

```bash
# Install WireGuard
apt install -y wireguard wireguard-tools

# WireGuard hanya perlu /dev/net/tun yang lebih mudah di-enable
```

## Verifikasi Setelah Setup

```bash
# Check devices
ls -la /dev/ppp /dev/net/tun

# Check kernel modules
lsmod | grep -E "ppp|l2tp|tun"

# Test PPP
pppd --version

# Test xl2tpd
systemctl status xl2tpd
```

## Troubleshooting

### PPP Interface Tidak Muncul (pppd exit code 2)

**Gejala:**
```bash
systemctl status xl2tpd
# Output:
# child_handler : pppd exited for call 1234 with code 2
# Connection disconnected immediately after connect
```

**Penyebab:** Authentication gagal karena CHAP credentials tidak tersedia

**Solusi:**

1. **Cek file chap-secrets**
   ```bash
   sudo cat /etc/ppp/chap-secrets
   ```
   
   Jika kosong atau hanya berisi comment, tambahkan credentials:
   ```bash
   sudo bash -c "cat > /etc/ppp/chap-secrets << 'EOF'
# Secrets for authentication using CHAP
# client        server  secret                  IP addresses
vpn-username * vpn-password *
EOF
"
   sudo chmod 600 /etc/ppp/chap-secrets
   ```

2. **Hapus duplicate settings di options file**
   ```bash
   sudo cat /etc/ppp/options.l2tpd.client
   ```
   
   **Pastikan TIDAK ADA baris berikut** (karena sudah di xl2tpd.conf):
   - `name vpn-username`
   - `password vpn-password`
   - `plugin pppol2tp.so`
   - `pppol2tp 7`
   
   File yang benar:
   ```bash
   sudo bash -c "cat > /etc/ppp/options.l2tpd.client << 'EOF'
ipcp-accept-local
ipcp-accept-remote
refuse-eap
require-mschap-v2
noccp
noauth
nodefaultroute
usepeerdns
debug
connect-delay 5000
EOF
"
   ```

3. **Restart service**
   ```bash
   sudo systemctl restart xl2tpd
   sleep 5
   ip addr show | grep -A3 ppp
   ```

4. **Cek hasil**
   - PPP interface harus UP (ppp0, ppp1, ppp2, dll)
   - Harus ada IP address assigned
   - Ping ke VPN gateway harus berhasil

### Error: "Couldn't open the /dev/ppp device: No such file or directory"

**Solusi**: Admin Proxmox perlu enable device seperti di Opsi 1

### Error: "Kernel doesn't support ppp_generic"

**Solusi**: Load kernel modules di Proxmox host:

```bash
# Di Proxmox host
modprobe ppp_generic
modprobe l2tp_ppp
```

### Error: "Operation not permitted" saat create device

**Solusi**: Devices tidak bisa dibuat dari dalam container, harus dari host Proxmox

## Contact Info untuk Admin Proxmox

Kirim request ke admin Proxmox dengan template berikut:

```
Subject: Request Enable TUN/TAP untuk Container ID XXX

Halo Admin,

Mohon bantuan untuk enable TUN/TAP devices untuk container RADIUS server.
Diperlukan untuk L2TP VPN client connection ke server pusat.

Container ID: XXX
Hostname: salfaradius

Command yang perlu dijalankan di Proxmox host:
```bash
CTID=XXX
pct set $CTID -features nesting=1,keyctl=1
pct set $CTID -dev0 /dev/net/tun,mode=0666
pct set $CTID -dev1 /dev/ppp,mode=0600
modprobe ppp_generic
modprobe l2tp_ppp
pct reboot $CTID
```

Terima kasih!
```

## Alternative: Gunakan VPS KVM/Dedicated Server

Jika admin Proxmox tidak bisa enable TUN/TAP, pertimbangkan untuk:
- Migrasi ke KVM VM (full virtualization)
- Gunakan VPS dedicated yang mendukung VPN
- Gunakan physical server

## References

- [Proxmox LXC TUN/TAP Setup](https://pve.proxmox.com/wiki/OpenVPN_in_LXC)
- [xl2tpd Documentation](https://github.com/xelerance/xl2tpd)
- [Linux PPP Kernel Module](https://www.kernel.org/doc/html/latest/networking/ppp_generic.html)
