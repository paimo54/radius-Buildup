# Fitur Import User PPPoE dari Excel/CSV

## Deskripsi
Fitur ini memungkinkan admin untuk mengimpor data user PPPoE dalam jumlah banyak (bulk import) menggunakan file Excel (.xlsx) atau CSV.

## File yang Diubah

### 1. Backend API
- **File**: `src/app/api/pppoe/users/bulk/route.ts`
- **Perubahan**:
  - Menambahkan library `xlsx` untuk membaca file Excel
  - Support download template dalam format Excel (.xlsx) dan CSV
  - Support upload dan parsing file Excel (.xlsx, .xls) dan CSV
  - Validasi dan import data user ke database
  - Sync otomatis ke RADIUS (radcheck dan radusergroup)

### 2. Frontend
- **File**: `src/app/admin/pppoe/users/page.tsx`
- **Perubahan**:
  - Tombol "Template Excel" untuk download template XLSX
  - Dialog import yang support file Excel dan CSV
  - Update file input untuk accept `.csv,.xlsx,.xls`
  - Informasi format file yang didukung

### 3. Dependencies
- **Package**: `xlsx` (sudah terinstall)
- Digunakan untuk membaca dan menulis file Excel

## Cara Menggunakan

### Download Template
1. Buka halaman PPPoE Users (`/admin/pppoe/users`)
2. Klik tombol **"Template Excel"** untuk download template dalam format XLSX
3. File `pppoe-users-template.xlsx` akan terdownload

### Format Template Excel
Template berisi kolom-kolom berikut:

| Kolom | Wajib | Deskripsi | Contoh |
|-------|-------|-----------|---------|
| username | Ya | Username PPPoE (unique) | user001 |
| password | Ya | Password PPPoE | pass123 |
| name | Ya | Nama lengkap pelanggan | John Doe |
| phone | Ya | Nomor telepon | 08123456789 |
| email | Tidak | Email pelanggan | john@example.com || area | Tidak | Nama area/wilayah pelanggan | Cluster A || address | Tidak | Alamat lengkap | Jl. Example No. 123 |
| ipAddress | Tidak | IP Address static | 10.10.10.2 |
| expiredAt | Tidak | Tanggal expired (YYYY-MM-DD) | 2025-12-31 |
| latitude | Tidak | Koordinat GPS latitude | -6.200000 |
| longitude | Tidak | Koordinat GPS longitude | 106.816666 |

### Import Data
1. Klik tombol **"Import"**
2. Dialog import akan terbuka
3. **Pilih File**: Upload file Excel (.xlsx, .xls) atau CSV
4. **Pilih Profile**: Pilih profile PPPoE untuk semua user yang diimport
5. **Pilih NAS** (opsional): Pilih router/NAS atau biarkan "Global"
6. Klik tombol **"Import"**
7. Sistem akan memproses dan menampilkan hasil:
   - Jumlah user berhasil diimport
   - Jumlah user gagal (jika ada)
   - Detail error untuk setiap baris yang gagal

## Validasi Import

### Validasi File
- File harus berformat CSV atau Excel (.xlsx, .xls)
- Minimal harus ada 1 baris data (selain header)
- Header harus mengandung kolom wajib: username, password, name, phone

### Validasi Data
Per baris data akan divalidasi:
1. **Field Wajib**: username, password, name, phone harus diisi
2. **Username Unique**: Username tidak boleh sudah ada di database
3. **Format Tanggal**: expiredAt harus format YYYY-MM-DD
4. **Format Koordinat**: latitude dan longitude harus berupa angka desimal

### Error Handling
Jika ada error, sistem akan:
- Tetap import data yang valid
- Skip data yang error
- Menampilkan detail error:
  - Nomor baris yang error
  - Username yang error
  - Pesan error

## Proses Import

### Alur Proses
1. **Upload File** → File Excel/CSV diupload ke server
2. **Parse File** → Data dibaca dan di-parse (xlsx untuk Excel, text parsing untuk CSV)
3. **Validasi** → Setiap baris data divalidasi
4. **Generate Customer ID** → Generate ID pelanggan unik (8 karakter)
5. **Create User** → Insert ke tabel `pppoe_users`
6. **Sync RADIUS** → Insert/update ke tabel `radcheck` dan `radusergroup`
7. **Result** → Tampilkan hasil import (success/failed count)

### Data yang Tersimpan
Untuk setiap user yang berhasil diimport:
- **pppoe_users**: Data user lengkap
- **radcheck**: Password untuk autentikasi RADIUS
- **radusergroup**: Group/profile untuk authorization RADIUS

## Keunggulan Format Excel

### Dibanding CSV:
1. **User-Friendly**: Lebih mudah diedit dengan Excel/LibreOffice
2. **Format Preserved**: Format tanggal, angka, dll tetap terjaga
3. **Styling**: Template bisa diberi warna, border, dll untuk panduan
4. **Multiple Sheets**: Bisa menambah sheet untuk dokumentasi/referensi
5. **Column Width**: Lebar kolom sudah diatur otomatis

### Template Excel Features:
- Header dengan nama kolom yang jelas
- 2 baris contoh data sebagai referensi
- Lebar kolom sudah dioptimalkan untuk kemudahan baca
- Format professional

## Testing

### Test Case 1: Import Excel Valid
- Upload file template.xlsx dengan data valid
- Semua user berhasil diimport
- Status "active" by default

### Test Case 2: Import dengan Username Duplicate
- Upload file dengan username yang sudah ada
- User duplicate akan di-skip
- User lain tetap diimport

### Test Case 3: Import CSV
- Upload file CSV dengan format sama
- Sistem tetap bisa membaca dan import data

### Test Case 4: Import dengan Field Opsional Kosong
- Email, address, ipAddress kosong
- Import tetap berhasil dengan nilai NULL

## Troubleshooting

### File tidak bisa dibaca
- Pastikan format file .xlsx, .xls, atau .csv
- Cek apakah file corrupt atau password-protected

### Import gagal semua
- Cek apakah header file sesuai template
- Pastikan ada data di baris kedua (bukan hanya header)

### Beberapa user gagal diimport
- Lihat detail error di hasil import
