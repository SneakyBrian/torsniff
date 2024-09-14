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

	// Use the first available device
	device := devices[0]

	// Create a WANIPConnection1 client
	clients, err := internetgateway1.NewWANIPConnection1ClientsFromRootDevice(device.Root, nil)
	if err != nil || len(clients) == 0 {
		return fmt.Errorf("error creating WANIPConnection1 client: %v", err)
	}
	client := clients[0]
	externalIP, err := client.GetExternalIPAddress()
	if err != nil {
		return fmt.Errorf("error getting external IP: %v", err)
	}

	log.Printf("External IP address: %s", externalIP)

	// Add port mapping
	localIP, err := getLocalIP()
	if err != nil {
		return fmt.Errorf("error getting local IP: %v", err)
	}

	err = client.AddPortMapping("", uint16(port), "UDP", uint16(port), localIP, true, "Torrent Indexer", 0)
	if err != nil {
		return fmt.Errorf("error adding port mapping: %v", err)
	}

	log.Printf("Port %d forwarded to local IP %s", port, localIP)
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
