package main

import (
	"fmt"
	"log"
	"net"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
)

type PortMapping struct {
	Port     int
	Protocol string // "TCP" or "UDP"
}

// SetupPortForwarding attempts to set up UPnP port forwarding rules for the specified ports and protocols.
func SetupPortForwarding(portMappings []PortMapping) error {
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

		// Iterate over all clients for each device
		for _, client := range clients {
			externalIP, err := client.GetExternalIPAddress()
			if err != nil {
				log.Printf("error getting external IP for device %s: %v", device.Location, err)
				continue
			}

			log.Printf("External IP address for device %s: %s", device.Location, externalIP)

			// Add port mapping for each port and protocol
			for _, mapping := range portMappings {
				err = client.AddPortMapping("", uint16(mapping.Port), mapping.Protocol, uint16(mapping.Port), localIP, true, "Torrent Indexer", 0)
				if err != nil {
					log.Printf("error adding port mapping for port %d (%s) on device %s: %v", mapping.Port, mapping.Protocol, device.Location, err)
					continue
				}

				log.Printf("Port %d (%s) forwarded to local IP %s on device %s", mapping.Port, mapping.Protocol, localIP, device.Location)
			}
		}
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
