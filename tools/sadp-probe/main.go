// sadp-probe - SADP discovery test tool
// Virtual kameralarni topish uchun multicast probe yuboradi va javoblarni ko'rsatadi.
//
// Ishlatish:
//   go run ./tools/sadp-probe/main.go
package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

const multicastAddr = "239.255.255.250:37020"

type ProbeRequest struct {
	XMLName      xml.Name `xml:"Probe"`
	Uuid         string   `xml:"Uuid"`
	Types        string   `xml:"Types"`
	ResponseMode int      `xml:"ResponseMode"`
}

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
	SoftwareVersion   string   `xml:"SoftwareVersion"`
}

func main() {
	addr, err := net.ResolveUDPAddr("udp4", multicastAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)
	_ = p.SetMulticastTTL(2)

	probe := ProbeRequest{
		Uuid:         fmt.Sprintf("probe-%d", time.Now().UnixNano()),
		Types:        "inquiry",
		ResponseMode: 2,
	}
	payload, _ := xml.Marshal(probe)
	payload = append([]byte(xml.Header), payload...)

	// Barcha interfeyslarga yuboramiz
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagMulticast == 0 {
			continue
		}
		if err := p.SetMulticastInterface(&iface); err != nil {
			continue
		}
		if _, err := p.WriteTo(payload, nil, addr); err != nil {
			log.Printf("Probe yuborilmadi (%s): %v", iface.Name, err)
		} else {
			log.Printf("Probe yuborildi: %s", iface.Name)
		}
	}

	fmt.Println("\nJavoblar kutilmoqda (10 soniya)...")
	fmt.Println("--------------------------------------")

	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 65535)
	found := 0

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		var resp ProbeResponse
		if err := xml.Unmarshal(buf[:n], &resp); err != nil {
			continue
		}

		found++
		fmt.Printf("[%d] Kamera topildi: %s\n", found, remoteAddr)
		fmt.Printf("    SN:          %s\n", resp.DeviceSN)
		fmt.Printf("    IP:          %s\n", resp.IPv4Address)
		fmt.Printf("    RTSP port:   %d\n", resp.CommandPort)
		fmt.Printf("    HTTP port:   %d\n", resp.HttpPort)
		fmt.Printf("    MAC:         %s\n", resp.MAC)
		fmt.Printf("    Type:        %s\n", resp.DeviceType)
		fmt.Printf("    Firmware:    %s\n", resp.SoftwareVersion)
		fmt.Printf("    RTSP URL:    rtsp://admin:A112233a@%s:%d/Streaming/Channels/101\n", resp.IPv4Address, resp.CommandPort)
		fmt.Printf("    ISAPI URL:   http://%s:%d/ISAPI/Streaming/channels/101/picture\n\n", resp.IPv4Address, resp.HttpPort)
	}

	fmt.Printf("--------------------------------------\n")
	fmt.Printf("Jami topilgan kameralar: %d ta\n", found)
}
