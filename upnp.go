package main

import (
	"fmt"
	"log"
	"net"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
)

// SetupPortForwarding attempts to set up a UPnP port forwarding rule for the specified port.
func SetupPortForwarding(port int) error {
	// Discover UPnP devices
	devices, err := goupnp.DiscoverDevices(internetgateway1.URN_WANIPConnection_1)
	if err != nil {
		return fmt.Errorf("error discovering devices: %v", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no UPnP devices found")
	}

	// Retrieve the local IP address
	localIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("error getting local IP: %v", err)
	}

	// Iterate over all discovered devices
	for _, device := range devices {
		// Create a WANIPConnection1 client for each device
		clients, err := internetgateway1.NewWANIPConnection1ClientsFromRootDevice(device.Root, nil)
		if err != nil || len(clients) == 0 {
			log.Printf("error creating WANIPConnection1 client for device %s: %v", device.Location, err)
			continue
		}

		client := clients[0]
		externalIP, err := client.GetExternalIPAddress()
		if err != nil {
			log.Printf("error getting external IP for device %s: %v", device.Location, err)
			continue
		}

		log.Printf("External IP address for device %s: %s", device.Location, externalIP)

		// Add port mapping for each device
		err = client.AddPortMapping("", uint16(port), "UDP", uint16(port), localIP, true, "Torrent Indexer", 0)
		if err != nil {
			log.Printf("error adding port mapping for device %s: %v", device.Location, err)
			continue
		}

		log.Printf("Port %d forwarded to local IP %s on device %s", port, localIP, device.Location)
	}
	return nil
}

// getLocalIP retrieves the local IP address of the machine.
func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
