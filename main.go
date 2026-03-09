package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/shranet/hikvision-virtual-cam/internal/config"
	"github.com/shranet/hikvision-virtual-cam/internal/isapi"
	"github.com/shranet/hikvision-virtual-cam/internal/rtsp"
	"github.com/shranet/hikvision-virtual-cam/internal/sadp"
)

func main() {
	imagesDir := flag.String("images", "./images", "Rasmlar joylashgan papka")
	basePort := flag.Int("base-port", 8554, "RTSP boshlang'ich porti (har bir kamera +1)")
	isapiPort := flag.Int("isapi-port", 8080, "ISAPI HTTP boshlang'ich porti (har bir kamera +1)")
	flag.Parse()

	images, err := findImages(*imagesDir)
	if err != nil || len(images) == 0 {
		log.Fatalf("Rasmlar topilmadi '%s' papkasida: %v", *imagesDir, err)
	}

	log.Printf("Topilgan rasmlar: %d ta", len(images))

	cameras := config.BuildCameras(images, *basePort, *isapiPort)
	for _, cam := range cameras {
		log.Printf("  Kamera %s -> RTSP:%d, HTTP:%d -> %s", cam.SN, cam.RTSPPort, cam.HttpPort, cam.ImagePath)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	rtspManager := rtsp.NewManager(cameras)
	go func() {
		if err := rtspManager.Start(ctx); err != nil {
			log.Printf("RTSP manager xatosi: %v", err)
		}
	}()

	isapiServer := isapi.NewServer(cameras)
	go func() {
		if err := isapiServer.Start(ctx); err != nil {
			log.Printf("ISAPI server xatosi: %v", err)
		}
	}()

	sadpServer := sadp.NewServer(cameras)
	go func() {
		if err := sadpServer.Start(ctx); err != nil {
			log.Printf("SADP server xatosi: %v", err)
		}
	}()

	fmt.Println("\n=== Virtual Hikvision Camera Server ===")
	fmt.Printf("SADP: UDP multicast 239.255.255.250:37020\n\n")
	for _, cam := range cameras {
		fmt.Printf("Kamera %s:\n", cam.SN)
		fmt.Printf("  RTSP:  rtsp://admin:A112233a@localhost:%d/Streaming/Channels/101\n", cam.RTSPPort)
		fmt.Printf("  ISAPI: http://localhost:%d/ISAPI/Streaming/channels/101/picture\n", cam.HttpPort)
		fmt.Printf("  Rasm:  %s\n\n", cam.ImagePath)
	}
	fmt.Println("To'xtatish uchun Ctrl+C bosing...")

	<-sig
	log.Println("To'xtatilmoqda...")
	cancel()
}

func findImages(dir string) ([]string, error) {
	exts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".bmp": true}
	var images []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if exts[ext] {
			images = append(images, filepath.Join(dir, e.Name()))
		}
	}
	return images, nil
}
