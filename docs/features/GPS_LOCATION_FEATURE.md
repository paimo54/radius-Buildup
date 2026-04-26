# Fitur Lokasi GPS untuk Registrasi Pelanggan

## Deskripsi
Fitur ini memungkinkan pelanggan yang mendaftar untuk menandai lokasi GPS mereka, dan admin dapat melihat lokasi tersebut melalui Google Maps.

## Perubahan yang Dilakukan

### 1. Database Schema
- **File**: `prisma/schema.prisma`
- **Perubahan**: Menambahkan field `latitude` dan `longitude` (Float, nullable) ke model `registrationRequest`

### 2. Migration
- **File**: `prisma/migrations/20251216_add_location_to_registration/migration.sql`
- **SQL**: 
  ```sql
  ALTER TABLE `registration_requests` 
  ADD COLUMN `latitude` DOUBLE NULL,
  ADD COLUMN `longitude` DOUBLE NULL;
  ```

### 3. Halaman Pendaftaran Pelanggan
- **File**: `src/app/daftar/page.tsx`
- **Fitur Baru**:
  - Input lokasi GPS (opsional)
  - Button "Pilih Lokasi di Peta" untuk membuka MapPicker
  - Menampilkan koordinat yang dipilih
  - Menggunakan komponen `MapPicker` yang sudah ada

### 4. API Endpoint
- **File**: `src/app/api/registrations/route.ts`
- **Perubahan**: Menerima dan menyimpan `latitude` dan `longitude` dari form pendaftaran

### 5. Halaman Admin
- **File**: `src/app/admin/pppoe/registrations/page.tsx`
- **Fitur Baru**:
  - Kolom "Lokasi" di tabel registrasi
  - Tombol "Lihat Lokasi" yang membuka Google Maps di tab baru
  - Menampilkan koordinat saat hover pada tombol
  - Jika tidak ada lokasi GPS, tampilkan "-"

## Cara Menggunakan

### Untuk Pelanggan
1. Buka halaman pendaftaran (`/daftar`)
2. Isi form pendaftaran seperti biasa
3. Klik tombol "Pilih Lokasi di Peta" (opsional)
4. Pilih lokasi dengan mengklik pada peta
5. Klik "Select Location" untuk menyimpan
6. Submit form pendaftaran

### Untuk Admin
1. Buka halaman Registrations (`/admin/pppoe/registrations`)
2. Di tabel, lihat kolom "Lokasi"
3. Jika pelanggan mengisi lokasi GPS, akan tampil tombol "Lihat Lokasi"
4. Klik tombol tersebut untuk membuka lokasi di Google Maps
5. Hover pada tombol untuk melihat koordinat lengkap

## Komponen yang Digunakan
- **MapPicker**: Komponen yang sudah ada di `src/components/MapPicker.tsx`
- Menggunakan Leaflet untuk menampilkan peta interaktif

---

## GPS Koordinat Clickable ke Google Maps (v2.11.6 — Phase 16)

Koordinat GPS di **tabel PPPoE Users** (`/admin/pppoe/users`) dapat diklik langsung untuk membuka lokasi di Google Maps di tab baru.

- Ditampilkan di kolom **Teknis** tabel PPPoE
- Jika pelanggan memiliki `latitude` dan `longitude`, tampil link yang bisa diklik
- Format URL: `https://www.google.com/maps?q={lat},{lng}`
- Jika tidak ada koordinat, tidak ada link yang tampil

Fitur ini berbeda dari GPS di halaman Registrasi (yang sudah ada sejak sebelumnya). GPS clickable di tabel PPPoE memudahkan admin melihat lokasi pelanggan aktif langsung dari halaman manajemen user.
- Support multiple basemap (street, satellite)
- Support marker drag untuk memilih lokasi

## Database Field
- **latitude**: DOUBLE NULL - Koordinat latitude lokasi pelanggan
- **longitude**: DOUBLE NULL - Koordinat longitude lokasi pelanggan
- Kedua field bersifat opsional (nullable)

## Link Google Maps
Format URL: `https://www.google.com/maps?q={latitude},{longitude}`
- Membuka di tab baru
- Langsung menampilkan marker di lokasi yang dipilih
