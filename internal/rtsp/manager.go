package rtsp

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/shranet/hikvision-virtual-cam/internal/config"
)

// Manager - barcha virtual kameralar uchun RTSP stream boshqaruvchi
type Manager struct {
	cameras []config.Camera
}

func NewManager(cameras []config.Camera) *Manager {
	return &Manager{cameras: cameras}
}

// Start - har bir kamera uchun alohida ffmpeg RTSP listen-mode stream ishga tushiradi
func (m *Manager) Start(ctx context.Context) error {
	checkDependencies()

	var wg sync.WaitGroup
	for _, cam := range m.cameras {
		wg.Add(1)
		go func(c config.Camera) {
			defer wg.Done()
			runFFmpegStream(ctx, c)
		}(cam)
	}

	wg.Wait()
	return nil
}

// runFFmpegStream - bitta kamera uchun rasmlarni 1fps da loop qilib RTSP stream qiladi.
// ffmpeg -rtsp_flags listen: server mode, klientni kutadi va ulanganida stream qiladi.
// Klient uzilgach ffmpeg chiqadi, loop uni qayta ishga tushiradi.
func runFFmpegStream(ctx context.Context, cam config.Camera) {
	rtspURL := fmt.Sprintf("rtsp://localhost:%d/Streaming/Channels/channels/%d", cam.RTSPPort, cam.Index)
	log.Printf("RTSP [%s]: Ishga tushmoqda -> %s (%d ta rasm)", cam.SN, rtspURL, len(cam.Images))

	// Concat faylni bir marta yaratamiz, funksiya chiqishida o'chiramiz
	concatFile, err := createConcatFile(cam)
	if err != nil {
		log.Printf("RTSP [%s]: concat fayl yaratishda xato: %v", cam.SN, err)
		return
	}
	defer os.Remove(concatFile)

	for {
		select {
		case <-ctx.Done():
			log.Printf("RTSP [%s]: To'xtatildi", cam.SN)
			return
		default:
		}

		// ffconcat + stream_loop orqali rasmlarni 1fps da cheksiz loop qiladi.
		// -g 1: har kadr keyframe (1fps uchun zarur).
		args := []string{
			"-re",
			"-stream_loop", "-1",
			"-f", "concat",
			"-safe", "0",
			"-i", concatFile,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-tune", "zerolatency",
			"-pix_fmt", "yuv420p",
			"-f", "rtsp",
			"-rtsp_transport", "tcp",
			rtspURL,
		}

		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("RTSP [%s]: Qayta ishga tushirilmoqda", cam.SN)
			}
		}
	}
}

// createConcatFile - ffconcat formatida vaqtinchalik fayl yaratadi.
// Har bir rasm 1 soniya davomida ko'rsatiladi.
func createConcatFile(cam config.Camera) (string, error) {
	f, err := os.CreateTemp("", fmt.Sprintf("hikvision_%s_*.txt", cam.SN))
	if err != nil {
		return "", err
	}
	defer f.Close()

	fmt.Fprintln(f, "ffconcat version 1.0")
	for _, img := range cam.Images {
		abs, err := filepath.Abs(img)
		if err != nil {
			abs = img
		}
		fmt.Fprintf(f, "file '%s'\n", abs)
		fmt.Fprintln(f, "duration 1")
	}
	return f.Name(), nil
}

func checkDependencies() {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Printf("OGOHLANTIRISH: 'ffmpeg' topilmadi => brew install ffmpeg")
	} else {
		log.Printf("OK: 'ffmpeg' topildi")
	}
}
