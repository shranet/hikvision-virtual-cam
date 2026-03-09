.PHONY: all build run deps clean install-deps test-sadp test-isapi test-rtsp help

# ── Sozlamalar ────────────────────────────────────────────────
BINARY    = hikvision-virtual-cam
IMAGES    = ./images
BASE_PORT = 8554
ISAPI_PORT = 8080

# ── Default ───────────────────────────────────────────────────
all: build

# ── Dependencies ─────────────────────────────────────────────
deps:
	go mod tidy
	go mod download

# ── macOS uchun kerakli toollarni o'rnatish ──────────────────
install-deps:
	@echo "==> ffmpeg o'rnatilmoqda..."
	brew install ffmpeg
	@echo "Barcha dependency tayyor."

# ── Build ─────────────────────────────────────────────────────
build: deps
	go build -o $(BINARY) .
	@echo "Build tayyor: ./$(BINARY)"

# ── Asosiy app ────────────────────────────────────────────────
run: build
	@echo "==> $(BINARY) ishga tushirilmoqda..."
	@echo "    Rasmlar: $(IMAGES)"
	@echo "    RTSP base port: $(BASE_PORT)"
	@echo "    ISAPI base port: $(ISAPI_PORT)"
	./$(BINARY) -images=$(IMAGES) -base-port=$(BASE_PORT) -isapi-port=$(ISAPI_PORT)

# ── Test ─────────────────────────────────────────────────────
test-sadp:
	@echo "SADP probe yuborilmoqda..."
	go run ./tools/sadp-probe/main.go

test-isapi:
	@echo "ISAPI rasm so'ralmoqda (kamera 1)..."
	curl -v -o /tmp/test_cam1.jpg "http://localhost:$(ISAPI_PORT)/ISAPI/Streaming/channels/101/picture"
	@echo "Rasm saqlandi: /tmp/test_cam1.jpg"

test-rtsp:
	@echo "RTSP stream tekshirilmoqda (kamera 1)..."
	ffprobe -v quiet -print_format json -show_streams "rtsp://localhost:$(BASE_PORT)/Streaming/Channels/101"

# ── Tozalash ─────────────────────────────────────────────────
clean:
	rm -f $(BINARY)
	pkill -f ffmpeg || true

# ── Yordam ───────────────────────────────────────────────────
help:
	@echo "Virtual Hikvision Camera - Komandalar:"
	@echo ""
	@echo "  make install-deps  - ffmpeg o'rnatish (macOS)"
	@echo "  make build         - Build"
	@echo "  make run           - App ishga tushirish"
	@echo "  make test-sadp     - SADP probe testi"
	@echo "  make test-isapi    - ISAPI rasm testi"
	@echo "  make test-rtsp     - RTSP stream testi"
	@echo "  make clean         - Tozalash"
	@echo ""
	@echo "  Parametrlar:"
	@echo "    BASE_PORT=8554   - RTSP boshlang'ich porti"
	@echo "    ISAPI_PORT=8080  - ISAPI boshlang'ich porti"
	@echo "    IMAGES=./images  - Rasmlar papkasi"
