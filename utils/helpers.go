package utils

import (
	"fmt"
	"net"
	"strings"
)

func LogError(message string) {
	fmt.Println("ERROR: " + message)
}

func LogDebug(message string) {
	fmt.Println("DEBUG: " + message)
}

func SanitizeFileName(fileName string) string {
	return strings.ReplaceAll(fileName, "..", "")
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "10.254.254.254:0")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func CalculateBroadcastAddress(ip, subnetMask string) string {
	ipAddr := net.ParseIP(ip).To4()
	mask := net.ParseIP(subnetMask).To4()

	if ipAddr == nil || mask == nil {
		LogError("Invalid IP or subnet mask")
		return "255.255.255.255"
	}

	broadcast := net.IPv4(
		ipAddr[0]|^mask[0],
		ipAddr[1]|^mask[1],
		ipAddr[2]|^mask[2],
		ipAddr[3]|^mask[3],
	)

	return broadcast.String()
}
