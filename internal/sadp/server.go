package sadp

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/shranet/hikvision-virtual-cam/internal/config"
	"golang.org/x/net/ipv4"
)

const (
	multicastAddr = "239.255.255.250:37020"
	listenPort    = 37020
)

// ProbeRequest - SADP Probe so'rovi
type ProbeRequest struct {
	XMLName      xml.Name `xml:"Probe"`
	Uuid         string   `xml:"Uuid"`
	Types        string   `xml:"Types"`
	ResponseMode int      `xml:"ResponseMode"`
}

// ProbeResponse - SADP Probe javobi (Hikvision formati)
type ProbeResponse struct {
	XMLName           xml.Name `xml:"ProbeMatch"`
	Uuid              string   `xml:"Uuid"`
	Types             string   `xml:"Types"`
	DeviceType        string   `xml:"DeviceType"`
	DeviceDescription string   `xml:"DeviceDescription"`
	DeviceSN          string   `xml:"DeviceSN"`
	CommandPort       int      `xml:"CommandPort"`
	HttpPort          int      `xml:"HttpPort"`
	MAC               string   `xml:"MAC"`
	IPv4Address       string   `xml:"IPv4Address"`
	IPv4SubnetMask    string   `xml:"IPv4SubnetMask"`
	IPv4Gateway       string   `xml:"IPv4Gateway"`
	IPv6Address       string   `xml:"IPv6Address"`
	DHCP              bool     `xml:"DHCP"`
	ChannelNumber     int      `xml:"ChannelNumber"`
	SoftwareVersion   string   `xml:"SoftwareVersion"`
}

// Server - SADP UDP multicast discovery server
type Server struct {
	cameras []config.Camera
}

func NewServer(cameras []config.Camera) *Server {
	return &Server{cameras: cameras}
}

func (s *Server) Start(ctx context.Context) error {
	laddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: listenPort,
	}

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return fmt.Errorf("SADP UDP listen xatosi: %w", err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)

	mcastIP := net.ParseIP("239.255.255.250")
	group := &net.UDPAddr{IP: mcastIP}

	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagMulticast == 0 {
			continue
		}
		if err := p.JoinGroup(&iface, group); err != nil {
			log.Printf("SADP: JoinGroup %s: %v", iface.Name, err)
		} else {
			log.Printf("SADP: Multicast guruhiga qo'shildi: %s", iface.Name)
		}
	}

	log.Printf("SADP server ishga tushdi - UDP :%d (multicast %s)", listenPort, multicastAddr)

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buf := make([]byte, 65535)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				continue
			}
		}

		data := buf[:n]

		var probe ProbeRequest
		if err := xml.Unmarshal(data, &probe); err != nil {
			continue
		}

		if !strings.Contains(strings.ToLower(probe.Types), "inquiry") {
			continue
		}

		log.Printf("SADP: Probe qabul qilindi %s dan (Uuid: %s)", remoteAddr, probe.Uuid)

		// Har bir virtual kamera uchun javob yuboramiz
		for _, cam := range s.cameras {
			resp := s.buildResponse(probe.Uuid, cam)
			respData, err := xml.Marshal(resp)
			if err != nil {
				continue
			}
			payload := append([]byte(xml.Header), respData...)

			time.Sleep(10 * time.Millisecond)

			_, err = conn.WriteToUDP(payload, remoteAddr)
			if err != nil {
				log.Printf("SADP: Javob xatosi %s: %v", cam.SN, err)
			} else {
				log.Printf("SADP: Javob -> %s (kamera: %s, rtsp:%d, http:%d)", remoteAddr, cam.SN, cam.RTSPPort, cam.HttpPort)
			}
		}
	}
}

func (s *Server) buildResponse(probeUuid string, cam config.Camera) ProbeResponse {
	return ProbeResponse{
		Uuid:              probeUuid,
		Types:             "HIK_DS-2CD2T47G2-L",
		DeviceType:        "IPC",
		DeviceDescription: fmt.Sprintf("Virtual Hikvision Camera #%d", cam.Index),
		DeviceSN:          cam.SN,
		CommandPort:       cam.RTSPPort,
		HttpPort:          cam.HttpPort,
		MAC:               cam.MAC,
		IPv4Address:       cam.IP,
		IPv4SubnetMask:    "255.255.255.0",
		IPv4Gateway:       "127.0.0.1",
		IPv6Address:       "",
		DHCP:              false,
		ChannelNumber:     1,
		SoftwareVersion:   "V5.7.15",
	}
}
