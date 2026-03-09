package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/shranet/hikvision-virtual-cam/internal/config"
	"github.com/shranet/hikvision-virtual-cam/internal/isapi"
	"github.com/shranet/hikvision-virtual-cam/internal/rtsp"
	"github.com/shranet/hikvision-virtual-cam/internal/sadp"
)

func main() {
	imagesDir := flag.String("images", "./images", "Kamera rasmlar papkasi (har bir kamera images/1/, images/2/, ...)")
	basePort := flag.Int("base-port", 8554, "mediamtx RTSP porti (barcha kameralar uchun bir xil)")
	isapiPort := flag.Int("isapi-port", 8080, "ISAPI HTTP boshlang'ich porti (har bir kamera +1)")
	flag.Parse()

	// images/1/, images/2/, ... papkalaridan kamera rasmlarini topamiz
	cameraDirs, err := findCameraDirs(*imagesDir)
	if err != nil || len(cameraDirs) == 0 {
		log.Fatalf("Kamera papkalari topilmadi '%s' ichida: %v\n"+
			"  Kutilayotgan tuzilma:\n"+
			"    images/1/photo1.jpg\n"+
			"    images/1/photo2.jpg\n"+
			"    images/2/photo1.jpg\n", *imagesDir, err)
	}

	cameras := config.BuildCameras(cameraDirs, *basePort, *isapiPort)
	for _, cam := range cameras {
		log.Printf("  Kamera %s -> RTSP:%d, HTTP:%d, rasmlar:%d (%s)",
			cam.SN, cam.RTSPPort, cam.HttpPort, len(cam.Images), cam.ImagesDir)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := rtsp.NewManager(cameras).Start(ctx); err != nil {
			log.Printf("RTSP manager xatosi: %v", err)
		}
	}()

	go func() {
		if err := isapi.NewServer(cameras).Start(ctx); err != nil {
			log.Printf("ISAPI server xatosi: %v", err)
		}
	}()

	go func() {
		if err := sadp.NewServer(cameras).Start(ctx); err != nil {
			log.Printf("SADP server xatosi: %v", err)
		}
	}()

	fmt.Println("\n=== Virtual Hikvision Camera Server ===")
	fmt.Printf("SADP: UDP multicast 239.255.255.250:37020\n\n")
	for _, cam := range cameras {
		fmt.Printf("Kamera %s (%d ta rasm):\n", cam.SN, len(cam.Images))
		fmt.Printf("  RTSP:  rtsp://localhost:%d/channels/%d\n", cam.RTSPPort, cam.Index)
		fmt.Printf("  ISAPI: http://localhost:%d/ISAPI/Streaming/channels/101/picture\n\n", cam.HttpPort)
	}
	fmt.Println("To'xtatish uchun Ctrl+C bosing...")

	<-sig
	log.Println("To'xtatilmoqda...")
	cancel()
}

// findCameraDirs - images/1/, images/2/, ... papkalarini numerik tartibda topadi.
// Har bir papka ichidagi rasmlar (jpg/png/bmp) saralangan holda qaytariladi.
func findCameraDirs(imagesDir string) ([][]string, error) {
	exts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".bmp": true}

	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, err
	}

	type dirEntry struct {
		num  int
		path string
	}
	var dirs []dirEntry

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // raqam bo'lmagan papkalarni o'tkazib yuboramiz
		}
		dirs = append(dirs, dirEntry{num: n, path: filepath.Join(imagesDir, e.Name())})
	}

	// Numerik tartibda saralash (1, 2, 3, ... 10, 11, ...)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].num < dirs[j].num
	})

	var result [][]string
	for _, d := range dirs {
		subEntries, err := os.ReadDir(d.path)
		if err != nil {
			continue
		}

		var images []string
		for _, se := range subEntries {
			if se.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(se.Name()))
			if exts[ext] {
				images = append(images, filepath.Join(d.path, se.Name()))
			}
		}

		if len(images) > 0 {
			result = append(result, images)
		}
	}

	return result, nil
}
