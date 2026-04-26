# Outage Notification System

Sistem untuk mengirim notifikasi gangguan/maintenance ke pelanggan melalui WhatsApp dan Email secara massal.

## Overview

Ketika terjadi gangguan jaringan atau ada jadwal maintenance, admin dapat mengirim notifikasi ke semua pelanggan yang terdampak dengan informasi:
- Jenis gangguan
- Deskripsi gangguan
- Estimasi waktu penyelesaian
- Area yang terdampak

## API Endpoint

**Endpoint**: `POST /api/pppoe/users/send-notification`

**Request Body**:
```json
{
  "userIds": [1, 2, 3, 4, 5],
  "notificationType": "outage",
  "notificationMethod": "whatsapp|email|both",
  "issueType": "Gangguan Jaringan",
  "description": "Terjadi gangguan pada backbone jaringan yang menyebabkan koneksi internet terputus.",
  "estimatedTime": "2-3 jam",
  "affectedArea": "Seluruh wilayah Kecamatan A"
}
```

### Parameter

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `userIds` | number[] | Yes | Array ID user PPPoE yang akan menerima notifikasi |
| `notificationType` | string | Yes | Harus "outage" untuk notifikasi gangguan |
| `notificationMethod` | string | Yes | "whatsapp", "email", atau "both" |
| `issueType` | string | Yes | Jenis gangguan (Gangguan Jaringan, Maintenance, Upgrade, Lainnya) |
| `description` | string | Yes | Deskripsi detail gangguan |
| `estimatedTime` | string | Yes | Estimasi waktu penyelesaian |
| `affectedArea` | string | Yes | Area yang terdampak |

### Issue Types

1. **Gangguan Jaringan** - Untuk masalah teknis seperti fiber cut, device down
2. **Maintenance** - Untuk jadwal maintenance rutin
3. **Upgrade** - Untuk upgrade jaringan/perangkat
4. **Lainnya** - Untuk jenis gangguan lain

### Response

**Success**:
```json
{
  "message": "Notifikasi gangguan berhasil dikirim",
  "totalSent": 50,
  "emailSent": 50,
  "whatsappSent": 50,
  "failedCount": 0
}
```

**Error**:
```json
{
  "error": "Gagal mengirim notifikasi",
  "failedCount": 2,
  "errors": [
    "User John: Nomor WhatsApp tidak valid",
    "User Jane: Email tidak terdaftar"
  ]
}
```

## WhatsApp Template

**Template Name**: `outage_notification`

**Template Content**:
```
⚠️ *PEMBERITAHUAN GANGGUAN JARINGAN*

Yth. Pelanggan *{{customerName}}*,
ID Pelanggan: {{customerId}}

Dengan ini kami informasikan bahwa sedang terjadi *{{issueType}}* pada jaringan kami.

📋 *Detail Gangguan:*
━━━━━━━━━━━━━━━━━━━━━━━━
📍 Area Terdampak: {{affectedArea}}
📝 Keterangan: {{description}}
⏱️ Estimasi Perbaikan: {{estimatedTime}}

Kami mohon maaf atas ketidaknyamanan ini. Tim teknisi kami sedang bekerja untuk menyelesaikan masalah ini secepat mungkin.

Untuk informasi lebih lanjut, silakan hubungi:
📞 {{companyPhone}}
📧 {{companyEmail}}

Terima kasih atas pengertian Anda.

Hormat kami,
*{{companyName}}*
```

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{customerName}}` | Nama pelanggan |
| `{{customerId}}` | ID pelanggan (generated) |
| `{{username}}` | Username PPPoE |
| `{{issueType}}` | Jenis gangguan |
| `{{description}}` | Deskripsi gangguan |
| `{{estimatedTime}}` | Estimasi waktu penyelesaian |
| `{{affectedArea}}` | Area yang terdampak |
| `{{companyName}}` | Nama ISP (dari settings) |
| `{{companyPhone}}` | Nomor telepon ISP |
| `{{companyEmail}}` | Email ISP |

## Email Template

**Template Name**: `outage_notification`

**Subject**: `[Penting] Pemberitahuan {{issueType}} - {{companyName}}`

**Body (HTML)**:
```html
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { background: #f44336; color: white; padding: 20px; text-align: center; }
    .content { padding: 20px; background: #f9f9f9; }
    .detail-box { background: white; padding: 15px; border-left: 4px solid #f44336; margin: 15px 0; }
    .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>⚠️ Pemberitahuan Gangguan Jaringan</h2>
    </div>
    <div class="content">
      <p>Yth. Pelanggan <strong>{{customerName}}</strong>,</p>
      <p>ID Pelanggan: {{customerId}}</p>
      
      <p>Dengan ini kami informasikan bahwa sedang terjadi <strong>{{issueType}}</strong> pada jaringan kami.</p>
      
      <div class="detail-box">
        <h4>Detail Gangguan:</h4>
        <ul>
          <li><strong>Area Terdampak:</strong> {{affectedArea}}</li>
          <li><strong>Keterangan:</strong> {{description}}</li>
          <li><strong>Estimasi Perbaikan:</strong> {{estimatedTime}}</li>
        </ul>
      </div>
      
      <p>Kami mohon maaf atas ketidaknyamanan ini. Tim teknisi kami sedang bekerja untuk menyelesaikan masalah ini secepat mungkin.</p>
      
      <p>Untuk informasi lebih lanjut, silakan hubungi:</p>
      <ul>
        <li>Telepon: {{companyPhone}}</li>
        <li>Email: {{companyEmail}}</li>
      </ul>
      
      <p>Terima kasih atas pengertian Anda.</p>
    </div>
    <div class="footer">
      <p>Hormat kami,<br><strong>{{companyName}}</strong></p>
    </div>
  </div>
</body>
</html>
```

## Usage Example (Frontend)

### React Component
```typescript
const sendOutageNotification = async () => {
  const response = await fetch('/api/pppoe/users/send-notification', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      userIds: selectedUsers.map(u => u.id),
      notificationType: 'outage',
      notificationMethod: notificationMethod,
      issueType: issueType,
      description: description,
      estimatedTime: estimatedTime,
      affectedArea: affectedArea
    })
  });
  
  const result = await response.json();
  
  if (result.error) {
    toast.error(result.error);
  } else {
    toast.success(`Notifikasi berhasil dikirim ke ${result.totalSent} pelanggan`);
  }
};
```

### Dialog Form
```typescript
<Dialog open={isOutageDialogOpen} onOpenChange={setIsOutageDialogOpen}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Kirim Notifikasi Gangguan</DialogTitle>
    </DialogHeader>
    
    <div className="space-y-4">
      <div>
        <Label>Jenis Gangguan</Label>
        <Select value={issueType} onValueChange={setIssueType}>
          <SelectTrigger>
            <SelectValue placeholder="Pilih jenis gangguan" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="Gangguan Jaringan">Gangguan Jaringan</SelectItem>
            <SelectItem value="Maintenance">Maintenance</SelectItem>
            <SelectItem value="Upgrade">Upgrade</SelectItem>
            <SelectItem value="Lainnya">Lainnya</SelectItem>
          </SelectContent>
        </Select>
      </div>
      
      <div>
        <Label>Deskripsi Gangguan</Label>
        <Textarea 
          value={description} 
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Jelaskan detail gangguan..."
        />
      </div>
      
      <div>
        <Label>Estimasi Waktu</Label>
        <Input 
          value={estimatedTime} 
          onChange={(e) => setEstimatedTime(e.target.value)}
          placeholder="contoh: 2-3 jam"
        />
      </div>
      
      <div>
        <Label>Area Terdampak</Label>
        <Input 
          value={affectedArea} 
          onChange={(e) => setAffectedArea(e.target.value)}
          placeholder="contoh: Seluruh area, atau Area A"
        />
      </div>
      
      <div>
        <Label>Metode Notifikasi</Label>
        <RadioGroup value={notificationMethod} onValueChange={setNotificationMethod}>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="whatsapp" id="wa" />
            <Label htmlFor="wa">WhatsApp</Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="email" id="email" />
            <Label htmlFor="email">Email</Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="both" id="both" />
            <Label htmlFor="both">WhatsApp & Email</Label>
          </div>
        </RadioGroup>
      </div>
    </div>
    
    <DialogFooter>
      <Button variant="outline" onClick={() => setIsOutageDialogOpen(false)}>
        Batal
      </Button>
      <Button onClick={sendOutageNotification}>
        Kirim Notifikasi
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

## Best Practices

1. **Select Target Users Carefully**
   - Filter by area jika hanya area tertentu yang terdampak
   - Filter by status Active untuk hanya kirim ke pelanggan aktif

2. **Provide Clear Information**
   - Deskripsi harus jelas dan mudah dipahami
   - Estimasi waktu realistis
   - Area terdampak spesifik

3. **Choose Appropriate Channel**
   - WhatsApp: Lebih cepat, real-time
   - Email: Lebih formal, ada record
   - Both: Untuk gangguan major

4. **Send Follow-up**
   - Kirim notifikasi update jika ada progress
   - Kirim notifikasi resolved ketika masalah selesai

## Troubleshooting

### Notifikasi tidak terkirim ke beberapa user
- Cek nomor WhatsApp valid (format: 62xxx)
- Cek email valid dan tidak kosong
- Cek WhatsApp provider status

### Error "Template not found"
- Pastikan template `outage_notification` sudah ada di database
- Jalankan seed untuk membuat template default

### Rate limiting
- Jika mengirim ke banyak user, mungkin terkena rate limit
- Sistem akan retry otomatis dengan delay
