# Virtual Hikvision Camera Server

A Go-based simulator that emulates real Hikvision IP cameras. Useful for testing NVR software, VMS integrations, or any system that needs to discover and stream from Hikvision cameras — without physical hardware.

Emulated protocols:
- **SADP** — UDP multicast device discovery (port 37020, group 239.255.255.250)
- **RTSP** — H.264 video stream via ffmpeg, looping through images at 1 fps
- **ISAPI** — HTTP snapshot endpoint returning images round-robin per request

---

## How It Works

### Image directories → Virtual cameras

Each subdirectory under `images/` becomes one virtual camera. The directory name becomes the camera's channel ID and must match `[0-9a-z]+`.

```
images/
├── 1/          → camera "virtualcam_1"     (RTSP channel: 1,     HTTP port: 8080)
├── 2/          → camera "virtualcam_2"     (RTSP channel: 2,     HTTP port: 8081)
└── lobby/      → camera "virtualcam_lobby" (RTSP channel: lobby, HTTP port: 8082)
```

Supported image formats: `.jpg`, `.jpeg`, `.png`, `.bmp`

### SADP discovery

The SADP server listens on UDP port 37020 and joins the multicast group `239.255.255.250` on all available network interfaces. When a client sends a `<Probe Types="inquiry">` XML packet, the server replies with a `<ProbeMatch>` response for each virtual camera, advertising:

- Device serial number (`virtualcam_{id}`)
- Emulated model: `HIK_DS-2CD2T47G2-L`, firmware `V5.7.15`
- Fake MAC address: `00:0C:29:AA:BB:{index}`
- RTSP and HTTP ports

### RTSP streaming

For each camera, the RTSP manager creates a temporary `ffconcat` playlist (each image shown for 1 second) and launches an `ffmpeg` process in RTSP server (listen) mode:

```
ffmpeg -re -stream_loop -1 -f concat -i <playlist>
       -c:v libx264 -preset veryfast -tune zerolatency -pix_fmt yuv420p
       -f rtsp -rtsp_transport tcp
       rtsp://localhost:{base-port}/Streaming/channels/{id}
```

If a client disconnects, ffmpeg exits and is automatically restarted.

### ISAPI snapshot

Each camera runs its own HTTP server on `base-isapi-port + index`. Every `GET /ISAPI/Streaming/channels/{id}/picture` request returns the next image in rotation (round-robin, independent of the RTSP stream).

Additional endpoint: `GET /ISAPI/System/deviceInfo` returns device metadata as XML.

---

## Requirements

- Go 1.22+
- ffmpeg (for RTSP streaming)

```bash
# macOS
brew install go ffmpeg

# Linux (Debian/Ubuntu)
sudo apt install golang ffmpeg
```

---

## Installation

```bash
git clone https://github.com/shranet/hikvision-virtual-cam
cd hikvision-virtual-cam
make build
```

---

## Usage

### 1. Prepare image directories

```bash
mkdir -p images/1 images/2
cp /path/to/photo1.jpg images/1/
cp /path/to/photo2.jpg images/1/
cp /path/to/other.jpg  images/2/
```

### 2. Run

```bash
# Default settings
make run

# Custom ports
./hikvision-virtual-cam \
    -images     ./images \
    -base-port  8554 \
    -isapi-port 8080
```

### 3. Output

```
=== Virtual Hikvision Camera Server ===
SADP: UDP multicast 239.255.255.250:37020

Camera virtualcam_1 (2 images):
  RTSP:  rtsp://localhost:8554/Streaming/channels/1
  ISAPI: http://localhost:8080/ISAPI/Streaming/channels/1/picture

Camera virtualcam_2 (1 images):
  RTSP:  rtsp://localhost:8554/Streaming/channels/2
  ISAPI: http://localhost:8081/ISAPI/Streaming/channels/2/picture
```

---

## Endpoints

### RTSP stream

```
rtsp://localhost:{base-port}/Streaming/channels/{id}
```

All cameras share the same RTSP port and are differentiated by channel path.

### ISAPI snapshot

```
http://localhost:{isapi-port + camera-index}/ISAPI/Streaming/channels/{id}/picture
```

Each request returns the next image in the camera's directory (cycles back to the first after the last).

| Camera dir   | HTTP port (default) | Snapshot URL |
|--------------|---------------------|--------------|
| `images/1/`  | 8080 | `http://localhost:8080/ISAPI/Streaming/channels/1/picture` |
| `images/2/`  | 8081 | `http://localhost:8081/ISAPI/Streaming/channels/2/picture` |
| `images/3/`  | 8082 | `http://localhost:8082/ISAPI/Streaming/channels/3/picture` |

### ISAPI device info

```
http://localhost:{isapi-port + camera-index}/ISAPI/System/deviceInfo
```

Returns device metadata (name, model, serial number, MAC, firmware) as XML.

---

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-images` | `./images` | Root directory containing camera subdirectories |
| `-base-port` | `8554` | RTSP server port (shared by all cameras, differentiated by channel path) |
| `-isapi-port` | `8080` | Base HTTP port; each camera gets `base + index` |

---

## Testing

### SADP discovery

```bash
# Send a multicast probe and print discovered cameras
go run ./tools/sadp-probe/main.go

# Or via Make
make test-sadp
```

### ISAPI snapshot

```bash
# Fetch snapshot from camera 1
curl -o /tmp/cam1.jpg http://localhost:8080/ISAPI/Streaming/channels/1/picture

# Fetch device info
curl http://localhost:8080/ISAPI/System/deviceInfo

# Make target (saves to /tmp/test_cam1.jpg)
make test-isapi
```

### RTSP stream

```bash
# Play in ffplay
ffplay rtsp://localhost:8554/Streaming/channels/1

# Inspect stream metadata
ffprobe -v quiet -print_format json -show_streams \
    rtsp://localhost:8554/Streaming/channels/1

# Play in VLC
vlc rtsp://localhost:8554/Streaming/channels/1

# Make target (runs ffprobe)
make test-rtsp
```

---

## Project Structure

```
hikvision-virtual-cam/
├── main.go                      # Entry point: parses flags, builds cameras, starts servers
├── internal/
│   ├── config/
│   │   └── config.go            # Camera struct and BuildCameras factory
│   ├── sadp/
│   │   └── server.go            # UDP multicast SADP discovery server
│   ├── rtsp/
│   │   └── manager.go           # ffmpeg-based RTSP stream manager
│   └── isapi/
│       └── server.go            # Per-camera ISAPI HTTP server
├── tools/
│   └── sadp-probe/
│       └── main.go              # CLI tool: sends SADP probe, prints discovered cameras
├── images/                      # Place your camera image directories here
├── Makefile
├── go.mod
└── go.sum
```

---

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile the binary |
| `make run` | Build and run with default settings |
| `make test-sadp` | Run SADP probe tool |
| `make test-isapi` | Fetch snapshot from camera 1 via curl |
| `make test-rtsp` | Inspect RTSP stream with ffprobe |
| `make install-deps` | Install ffmpeg via Homebrew (macOS) |
| `make clean` | Remove binary and kill any leftover ffmpeg processes |

---

## Troubleshooting

### Port already in use

```bash
./hikvision-virtual-cam -base-port 9554 -isapi-port 9080
```

### No cameras found at startup

The app expects at least one subdirectory under `images/` containing image files:

```
images/
└── 1/
    └── photo.jpg    ← at least one image required
```

Directory names must match `[0-9a-z]+` — no uppercase, no spaces, no special characters.

### SADP not receiving responses (macOS)

macOS firewall may block incoming UDP packets. Go to **System Settings → Network → Firewall** and allow incoming connections for the binary, or temporarily disable the firewall for testing.

### RTSP stream does not connect

Ensure ffmpeg is installed and accessible in `$PATH`. The stream uses TCP transport (`-rtsp_transport tcp`) which is more reliable than UDP over loopback.
