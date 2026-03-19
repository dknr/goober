package wol

import (
	"fmt"
	"net"
)

const (
	magicPacketRepeat = 16
)

var (
	magicPacketHeader = [6]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

// SendMagicPacket sends a Wake-on-LAN magic packet to the specified MAC address
// Broadcast address is typically 255.255.255.255 on port 9
func SendMagicPacket(mac string) error {
	// Parse MAC address
	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("invalid MAC address %s: %w", mac, err)
	}

	// Create magic packet (102 bytes: 6 header + 16 * 6 MAC)
	packet := make([]byte, 0, 102)
	for i := 0; i < magicPacketRepeat; i++ {
		packet = append(packet, magicPacketHeader[:]...)
		packet = append(packet, macAddr...)
	}

	// Broadcast to port 9
	addr := &net.UDPAddr{
		IP:   net.ParseIP("255.255.255.255"),
		Port: 9,
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send WoL packet: %w", err)
	}

	return nil
}

// SendMagicPacketTo sends a Wake-on-LAN magic packet to the specified MAC address
// on the specified network interface
func SendMagicPacketTo(mac, interfaceName string) error {
	// Parse MAC address
	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("invalid MAC address %s: %w", mac, err)
	}

	// Get network interface
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %w", interfaceName, err)
	}

	// Get all IP addresses for this interface
	addrs, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("failed to get interface addresses: %w", err)
	}

	// Find the first IPv4 address
	var ipAddr *net.IPNet
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				ipAddr = ipnet
				break
			}
		}
	}

	if ipAddr == nil {
		return fmt.Errorf("no IPv4 address found on interface %s", interfaceName)
	}

	// Create magic packet
	packet := make([]byte, 0, 102)
	for i := 0; i < magicPacketRepeat; i++ {
		packet = append(packet, magicPacketHeader[:]...)
		packet = append(packet, macAddr...)
	}

	// Send to the specific broadcast address for this network
	broadcastAddr := &net.UDPAddr{
		IP:   make(net.IP, 4),
		Port: 9,
	}
	copy(broadcastAddr.IP, ipAddr.IP.To4())
	broadcastAddr.IP[3] = 255 // Broadcast to subnet

	// Create UDP connection using the specific interface
	conn, err := net.DialUDP("udp", nil, broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send WoL packet: %w", err)
	}

	return nil
}