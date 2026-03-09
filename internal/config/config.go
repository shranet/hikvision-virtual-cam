package config

import "fmt"

// CameraDir - kamera papkasi ma'lumotlari (ID = papka nomi)
type CameraDir struct {
	ID     string
	Images []string
}

// Camera - bitta virtual kamera konfiguratsiyasi
type Camera struct {
	SN        string   // virtualcam_1, virtualcam_2, ...
	ID        string   // papka nomi: "1", "2", "cam1", ...
	ImagesDir string   // images/1, images/2, ...
	Images    []string // rasmlar ro'yxati (saralangan)
	RTSPPort  int      // mediamtx RTSP porti (bir xil, barcha kameralar uchun)
	HttpPort  int      // ISAPI HTTP porti (8080, 8081, ...)
	IP        string   // har doim "127.0.0.1"
	MAC       string   // fake MAC
}

// BuildCameras - CameraDir ro'yxatidan kameralar ro'yxati yasaydi.
func BuildCameras(cameraDirs []CameraDir, baseRTSPPort, baseHttpPort int) []Camera {
	cameras := make([]Camera, len(cameraDirs))
	for i, d := range cameraDirs {
		cameras[i] = Camera{
			SN:        fmt.Sprintf("virtualcam_%s", d.ID),
			ID:        d.ID,
			ImagesDir: fmt.Sprintf("images/%s", d.ID),
			Images:    d.Images,
			RTSPPort:  baseRTSPPort,
			HttpPort:  baseHttpPort + i,
			IP:        "127.0.0.1",
			MAC:       fmt.Sprintf("00:0C:29:AA:BB:%02X", i+1),
		}
	}
	return cameras
}
