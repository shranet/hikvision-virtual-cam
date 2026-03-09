# Virtual Hikvision Camera Server

Go tilida yozilgan virtual Hikvision kamera simulyatori. Haqiqiy Hikvision kamera protokollarini emulatsiya qiladi:

- **SADP** — UDP multicast orqali kameralarni discovery
- **RTSP** — ffmpeg orqali rasm fayllarini video stream sifatida uzatish
- **ISAPI** — HTTP orqali snapshot (rasm) qaytarish

---

## O'rnatish

### Talablar

- Go 1.21+
- ffmpeg (RTSP stream uchun)
- *(ixtiyoriy)* mediamtx — agar ffmpeg listen mode ishlashida muammo bo'lsa

```bash
# macOS
brew install go
brew install ffmpeg

# mediamtx (ixtiyoriy)
brew install mediamtx
```

### Loyihani yuklab olish

```bash
git clone <repo>
cd hikvision-virtual-cam
make deps
make build
```

---

## Ishlatish

### 1. Rasmlar papkasini tayyorlang

```bash
mkdir images
cp /path/to/your/photo1.jpg images/
cp /path/to/your/photo2.jpg images/
cp /path/to/your/photo3.png images/
```

Har bir rasm = 1 ta virtual kamera.

### 2. Ishga tushiring

```bash
# Default sozlamalar bilan
make run

# Yoki parametrlar bilan
./hikvision-virtual-cam \
    -images ./images \
    -base-port 8554 \
    -isapi-port 8080
```

### 3. Natija

```
=== Virtual Hikvision Camera Server ===
ISAPI: http://localhost:8080
SADP: UDP multicast 239.255.255.250:37020

Kamera virtualcam_1: rtsp://admin:A112233a@localhost:8554/Streaming/Channels/101
                     http://localhost:8080/ISAPI/Streaming/channels/101/picture

Kamera virtualcam_2: rtsp://admin:A112233a@localhost:8555/Streaming/Channels/101
                     http://localhost:8080/ISAPI/Streaming/channels/201/picture
```

---

## Testlash

### SADP Discovery testi

```bash
# Alohida terminal ochib:
go run ./tools/sadp-probe/main.go

# Yoki:
make test-sadp
```

### ISAPI rasm testi

```bash
# Birinchi kamera rasmi
curl -o /tmp/cam1.jpg "http://localhost:8080/ISAPI/Streaming/channels/101/picture"

# Ikkinchi kamera rasmi
curl -o /tmp/cam2.jpg "http://localhost:8080/ISAPI/Streaming/channels/201/picture"

# SN orqali
curl -o /tmp/cam1.jpg "http://localhost:8080/ISAPI/Streaming/channels/101/picture?sn=virtualcam_1"

# Index orqali
curl -o /tmp/cam1.jpg "http://localhost:8080/ISAPI/Streaming/channels/101/picture?index=1"
```

### RTSP stream testi

```bash
# ffplay orqali ko'rish
ffplay rtsp://localhost:8554/Streaming/Channels/101

# ffprobe orqali info
ffprobe -v quiet -print_format json -show_streams rtsp://localhost:8554/Streaming/Channels/101

# VLC orqali
vlc rtsp://localhost:8554/Streaming/Channels/101
```

---

## URL Formatlari

### RTSP
```
rtsp://admin:A112233a@localhost:{BASE_PORT + camera_index - 1}/Streaming/Channels/101

Misol:
  Kamera 1: rtsp://admin:A112233a@localhost:8554/Streaming/Channels/101
  Kamera 2: rtsp://admin:A112233a@localhost:8555/Streaming/Channels/101
  Kamera 3: rtsp://admin:A112233a@localhost:8556/Streaming/Channels/101
```

### ISAPI Picture
```
http://localhost:{ISAPI_PORT}/ISAPI/Streaming/channels/{CHANNEL_ID}/picture

CHANNEL_ID = kamera_index * 100 + 1
  Kamera 1: /ISAPI/Streaming/channels/101/picture
  Kamera 2: /ISAPI/Streaming/channels/201/picture
  Kamera 3: /ISAPI/Streaming/channels/301/picture

Parametrlar bilan:
  ?sn=virtualcam_1   (serial number orqali)
  ?index=1           (index orqali)
```

### Kameralar ro'yxati
```
http://localhost:8080/cameras
```

---

## Parametrlar

| Parametr | Default | Ta'rif |
|----------|---------|--------|
| `-images` | `./images` | Rasmlar papkasi |
| `-base-port` | `8554` | Birinchi kamera RTSP porti |
| `-isapi-port` | `8080` | ISAPI HTTP server porti |

---

## Loyiha tuzilmasi

```
hikvision-virtual-cam/
├── cmd/
│   └── main.go              # Asosiy entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Kamera konfiguratsiyasi
│   ├── sadp/
│   │   └── server.go        # SADP UDP discovery server
│   ├── rtsp/
│   │   └── manager.go       # RTSP stream manager (ffmpeg)
│   └── isapi/
│       └── server.go        # ISAPI HTTP server
├── tools/
│   └── sadp-probe/
│       └── main.go          # SADP test tool
├── images/                  # Rasmlar (siz qo'shasiz)
├── mediamtx.yml             # mediamtx konfiguratsiyasi
├── Makefile
└── README.md
```

---

## Muammolar va yechimlar

### ffmpeg "listen" mode ishlashmasa

`-rtsp_flags listen` ba'zi versiyalarda ishlamasligi mumkin. Bu holda **mediamtx** ishlating:

```bash
# Terminal 1: mediamtx
mediamtx mediamtx.yml

# Terminal 2: app (push mode)
# manager.go da listen flag ni olib, push mode ga o'zgartiring
```

### Port band bo'lsa

```bash
# Boshqa portni ishlating
./hikvision-virtual-cam -base-port 9554
```

### macOS firewall muammosi

SADP UDP multicast uchun firewall ruxsat berishi kerak. System Preferences → Security → Firewall → Allow incoming connections.
