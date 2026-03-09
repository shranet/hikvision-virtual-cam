package isapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/shranet/hikvision-virtual-cam/internal/config"
)

// Server - har bir virtual kamera uchun alohida ISAPI HTTP server
type Server struct {
	cameras []config.Camera
}

func NewServer(cameras []config.Camera) *Server {
	return &Server{cameras: cameras}
}

// Start - har bir kamera uchun alohida HTTP server ishga tushiradi
func (s *Server) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, cam := range s.cameras {
		wg.Add(1)
		go func(c config.Camera) {
			defer wg.Done()
			s.startCameraServer(ctx, c)
		}(cam)
	}

	wg.Wait()
	return nil
}

func (s *Server) startCameraServer(ctx context.Context, cam config.Camera) {
	mux := http.NewServeMux()

	// /ISAPI/Streaming/channels/101/picture - rasmni qaytaradi
	mux.HandleFunc("/ISAPI/Streaming/channels/101/picture", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(cam.ImagePath)
		if err != nil {
			http.Error(w, "image not found", http.StatusNotFound)
			return
		}

		ext := strings.ToLower(filepath.Ext(cam.ImagePath))
		contentType := "image/jpeg"
		switch ext {
		case ".png":
			contentType = "image/png"
		case ".bmp":
			contentType = "image/bmp"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	// /ISAPI/System/deviceInfo - kamera ma'lumotlari
	mux.HandleFunc("/ISAPI/System/deviceInfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<DeviceInfo>
  <deviceName>Virtual Hikvision Camera #%d</deviceName>
  <deviceID>%s</deviceID>
  <model>DS-2CD2T47G2-L</model>
  <serialNumber>%s</serialNumber>
  <macAddress>%s</macAddress>
  <firmwareVersion>V5.7.15</firmwareVersion>
</DeviceInfo>`, cam.Index, cam.SN, cam.SN, cam.MAC)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cam.HttpPort),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	log.Printf("ISAPI [%s]: http://localhost:%d/ISAPI/Streaming/channels/101/picture", cam.SN, cam.HttpPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("ISAPI [%s]: server xatosi: %v", cam.SN, err)
	}
}
