package config

import "fmt"

// Camera - bitta virtual kamera konfiguratsiyasi
type Camera struct {
	SN        string   // virtualcam_1, virtualcam_2, ...
	Index     int      // 1, 2, 3, ...
	ImagesDir string   // images/1, images/2, ...
	Images    []string // rasmlar ro'yxati (saralangan)
	RTSPPort  int      // mediamtx RTSP porti (bir xil, barcha kameralar uchun)
	HttpPort  int      // ISAPI HTTP porti (8080, 8081, ...)
	IP        string   // har doim "127.0.0.1"
	MAC       string   // fake MAC
}

// BuildCameras - papkalar ro'yxatidan kameralar ro'yxati yasaydi.
// cameraDirs[i] = i-kamera uchun rasmlar ro'yxati (images/1/, images/2/, ...).
func BuildCameras(cameraDirs [][]string, baseRTSPPort, baseHttpPort int) []Camera {
	cameras := make([]Camera, len(cameraDirs))
	for i, imgs := range cameraDirs {
		idx := i + 1
		dir := ""
		if len(imgs) > 0 {
			dir = fmt.Sprintf("images/%d", idx)
		}
		cameras[i] = Camera{
			SN:        fmt.Sprintf("virtualcam_%d", idx),
			Index:     idx,
			ImagesDir: dir,
			Images:    imgs,
			RTSPPort:  baseRTSPPort,
			HttpPort:  baseHttpPort + i,
			IP:        "127.0.0.1",
			MAC:       fmt.Sprintf("00:0C:29:AA:BB:%02X", idx),
		}
	}
	return cameras
}
