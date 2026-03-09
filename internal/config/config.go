package config

import "fmt"

// Camera - bitta virtual kamera konfiguratsiyasi
type Camera struct {
	SN        string // virtualcam_1, virtualcam_2, ...
	Index     int    // 1, 2, 3, ...
	ImagePath string // rasm fayl yo'li
	RTSPPort  int    // RTSP stream porti (8554, 8555, ...)
	HttpPort  int    // ISAPI HTTP porti (8080, 8081, ...)
	IP        string // har doim "127.0.0.1"
	MAC       string // fake MAC
}

// BuildCameras - rasmlar ro'yxatidan kameralar ro'yxati yasaydi
func BuildCameras(images []string, baseRTSPPort, baseHttpPort int) []Camera {
	cameras := make([]Camera, len(images))
	for i, img := range images {
		idx := i + 1
		cameras[i] = Camera{
			SN:        fmt.Sprintf("virtualcam_%d", idx),
			Index:     idx,
			ImagePath: img,
			RTSPPort:  baseRTSPPort + i,
			HttpPort:  baseHttpPort + i,
			IP:        "127.0.0.1",
			MAC:       fmt.Sprintf("00:0C:29:AA:BB:%02X", idx),
		}
	}
	return cameras
}
