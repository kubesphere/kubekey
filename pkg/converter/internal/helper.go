package internal

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
)

// parseIP parses a CIDR or an IP range string (e.g., "xxx-xxx") into a slice of actual IPs.
func parseIP(ip string) []string {
	var availableIPs []string

	// Trim trailing slash from IP if present
	ip = strings.TrimRight(ip, "/")
	if strings.Contains(ip, "/") {
		if strings.HasSuffix(ip, "/32") {
			// Handle single IP case in CIDR format
			availableIPs = append(availableIPs, strings.TrimSuffix(ip, "/32"))
		} else {
			// Parse CIDR into available IPs
			availableIPs = getAvailableIP(ip)
		}
	} else if strings.Contains(ip, "-") {
		// Handle IP range format (e.g., "192.168.1.1-192.168.1.10")
		ipRange := strings.SplitN(ip, "-", 2)
		if len(ipRange) == 2 {
			availableIPs = getAvailableIPRange(ipRange[0], ipRange[1])
		}
	} else {
		// Single IP case
		availableIPs = append(availableIPs, ip)
	}

	return availableIPs
}

// getAvailableIPRange generates all IPs between the start and end IP addresses.
func getAvailableIPRange(ipStart, ipEnd string) []string {
	var availableIPs []string

	startIP := net.ParseIP(ipStart).To4()
	endIP := net.ParseIP(ipEnd).To4()
	if startIP == nil || endIP == nil {
		return availableIPs
	}

	startIPNum := ipToInt(startIP)
	endIPNum := ipToInt(endIP)

	for ipNum := startIPNum; ipNum <= endIPNum; ipNum++ {
		availableIPs = append(availableIPs, intToIP(ipNum).String())
	}

	return availableIPs
}

// getAvailableIP calculates all available IPs in a given CIDR.
func getAvailableIP(ipAndMask string) []string {
	var availableIPs []string

	// Ensure the input is in CIDR format
	ipAndMask = strings.TrimSpace(ipAndMask)
	ipAndMask = iPAddressToCIDR(ipAndMask)

	_, ipnet, err := net.ParseCIDR(ipAndMask)
	if err != nil || ipnet == nil {
		return availableIPs
	}

	firstIP, lastIP := networkRange(ipnet)
	startIPNum := ipToInt(firstIP)
	endIPNum := ipToInt(lastIP)

	// Exclude the network and broadcast addresses
	for ipNum := startIPNum + 1; ipNum < endIPNum; ipNum++ {
		availableIPs = append(availableIPs, intToIP(ipNum).String())
	}

	return availableIPs
}

// ipToInt converts an IP address to a uint32.
func ipToInt(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

// intToIP converts a uint32 to an IP address.
func intToIP(n uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)

	return net.IP(b)
}

// iPAddressToCIDR converts an IP address with a subnet mask to CIDR format.
func iPAddressToCIDR(ipAddress string) string {
	if strings.Contains(ipAddress, "/") {
		parts := strings.Split(ipAddress, "/")
		ip := parts[0]
		mask := parts[1]
		if strings.Contains(mask, ".") {
			mask = iPMaskStringToCIDR(mask)
		}

		return ip + "/" + mask
	}

	return ipAddress
}

// iPMaskStringToCIDR converts a subnet mask string (e.g., "255.255.255.0") to a CIDR prefix length.
func iPMaskStringToCIDR(netmask string) string {
	parts := strings.Split(netmask, ".")
	if len(parts) != 4 {
		return "0"
	}

	maskBytes := make([]byte, 4)
	for i, part := range parts {
		val, _ := strconv.Atoi(part)
		maskBytes[i] = byte(val)
	}

	mask := net.IPv4Mask(maskBytes[0], maskBytes[1], maskBytes[2], maskBytes[3])
	ones, _ := mask.Size()

	return strconv.Itoa(ones)
}

// networkRange calculates the first and last IP in a given network.
func networkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP.To4()
	if netIP == nil {
		return nil, nil
	}

	startIP := netIP.Mask(network.Mask)
	endIP := make(net.IP, len(startIP))
	for i := range startIP {
		endIP[i] = startIP[i] | ^network.Mask[i]
	}

	return startIP, endIP
}
