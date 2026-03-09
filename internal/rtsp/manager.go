package rtsp

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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

// Start - barcha kameralar uchun ffmpeg RTSP listen mode da ishga tushiradi
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

// runFFmpegStream - bitta kamera uchun ffmpeg RTSP server sifatida ishlatadi
// ffmpeg -rtsp_flags listen bilan klientni kutadi va stream qiladi
func runFFmpegStream(ctx context.Context, cam config.Camera) {
	rtspURL := fmt.Sprintf("rtsp://localhost:%d/Streaming/Channels/101", cam.RTSPPort)
	log.Printf("RTSP [%s]: Ishga tushmoqda -> %s", cam.SN, rtspURL)

	for {
		select {
		case <-ctx.Done():
			log.Printf("RTSP [%s]: To'xtatildi", cam.SN)
			return
		default:
		}

		// Still image (jpg/png) ni loop qilib RTSP server sifatida stream qiladi.
		// -loop 1         : rasmni cheksiz takrorlaydi
		// -framerate 10   : 10fps kirishda o'qiydi
		// -rtsp_flags listen : server mode, klientni kutadi
		args := []string{
			"-re",
			"-loop", "1",
			"-framerate", "1",
			"-i", cam.ImagePath,
			"-vf", "fps=1,format=yuv420p",
			"-c:v", "libx264",
			"-tune", "stillimage",
			"-preset", "ultrafast",
			"-b:v", "500k",
			"-g", "20",
			"-f", "rtsp",
			"-rtsp_flags", "listen",
			rtspURL,
		}

		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("RTSP [%s]: Qayta ishga tushirilmoqda (xato: %v)", cam.SN, err)
			}
		}
	}
}

func checkDependencies() {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Printf("OGOHLANTIRISH: 'ffmpeg' topilmadi => brew install ffmpeg")
	} else {
		log.Printf("OK: 'ffmpeg' topildi")
	}
}
