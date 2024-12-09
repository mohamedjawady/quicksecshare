package networking

import (
	"fmt"
	"net"
	"time"

	"github.com/mohamedhabib/file-sharing-app/utils"
)

func AdvertiseService(servicePort string) {

	localIP := utils.GetLocalIP()
	if localIP == "" {
		utils.LogError("Could not determine local IP")
		return
	}

	broadcastIP := utils.CalculateBroadcastAddress(localIP, "255.255.255.0") // Adjust subnet mask if needed

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", broadcastIP, "9999"))
	if err != nil {
		utils.LogError("Error resolving UDP broadcast address: " + err.Error())
		return
	}

	conn, err := net.ListenPacket("udp4", fmt.Sprintf("%s:0", localIP))
	if err != nil {
		utils.LogError("Error listening on UDP: " + err.Error())
		return
	}
	defer conn.Close()

	utils.LogDebug("Starting to advertise service on port " + servicePort)

	for {
		message := []byte(fmt.Sprintf("SERVICE:FileExchanger:%s:%s", localIP, servicePort))
		_, err := conn.WriteTo(message, addr)
		if err != nil {
			utils.LogError("Error sending UDP broadcast: " + err.Error())
			return
		}

		utils.LogDebug(fmt.Sprintf("Service advertised: %s", message))
		time.Sleep(2 * time.Second)
	}
}
