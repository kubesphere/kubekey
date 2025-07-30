package _const

import (
	"encoding/binary"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
)

// ===========================================================================
// =============================   ParseIP   =================================
// ===========================================================================

// ParseIP parses a CIDR, an IP range string (e.g., "xxx-xxx"), or a single IP into a slice of actual IPs.
// Supports both IPv4 and IPv6.
func ParseIP(ip string) []string {
	var availableIPs []string

	ip = strings.TrimRight(ip, "/")
	if strings.Contains(ip, "/") {
		// Handle single IP case in CIDR format
		if strings.HasSuffix(ip, "/32") || strings.HasSuffix(ip, "/128") {
			availableIPs = append(availableIPs, strings.Split(ip, "/")[0])
		} else {
			availableIPs = getAvailableIP(ip)
		}
	} else if strings.Contains(ip, "-") {
		ipRange := strings.SplitN(ip, "-", 2)
		if len(ipRange) == 2 {
			availableIPs = getAvailableIPRange(ipRange[0], ipRange[1])
		}
	} else {
		availableIPs = append(availableIPs, ip)
	}

	return availableIPs
}

// getAvailableIPRange generates all IPs between the start and end IP addresses (inclusive).
// Supports both IPv4 and IPv6.
func getAvailableIPRange(ipStart, ipEnd string) []string {
	var availableIPs []string

	startIP := net.ParseIP(ipStart)
	endIP := net.ParseIP(ipEnd)
	if startIP == nil || endIP == nil {
		return availableIPs
	}

	// Determine if IPv4 or IPv6
	if startIP.To4() != nil && endIP.To4() != nil {
		startIP = startIP.To4()
		endIP = endIP.To4()
		startIPNum := ip4ToInt(startIP)
		endIPNum := ip4ToInt(endIP)
		for ipNum := new(big.Int).Set(startIPNum); ipNum.Cmp(endIPNum) <= 0; ipNum.Add(ipNum, big.NewInt(1)) {
			availableIPs = append(availableIPs, intToIP4(ipNum).String())
		}
	} else if startIP.To16() != nil && endIP.To16() != nil {
		startIP = startIP.To16()
		endIP = endIP.To16()
		startIPNum := ip6ToInt(startIP)
		endIPNum := ip6ToInt(endIP)
		for ipNum := new(big.Int).Set(startIPNum); ipNum.Cmp(endIPNum) <= 0; ipNum.Add(ipNum, big.NewInt(1)) {
			availableIPs = append(availableIPs, intToIP6(ipNum).String())
		}
	}

	return availableIPs
}

// getAvailableIP calculates all available IPs in a given CIDR.
// Supports both IPv4 and IPv6.
func getAvailableIP(ipAndMask string) []string {
	var availableIPs []string

	ipAndMask = strings.TrimSpace(ipAndMask)
	ipAndMask = iPAddressToCIDR(ipAndMask)

	_, ipnet, err := net.ParseCIDR(ipAndMask)
	if err != nil || ipnet == nil {
		return availableIPs
	}

	firstIP, lastIP := networkRange(ipnet)
	if firstIP == nil || lastIP == nil {
		return availableIPs
	}

	// IPv4
	if firstIP.To4() != nil {
		startIPNum := ip4ToInt(firstIP)
		endIPNum := ip4ToInt(lastIP)
		// Exclude network and broadcast addresses if possible
		for ipNum := new(big.Int).Add(startIPNum, big.NewInt(1)); ipNum.Cmp(endIPNum) < 0; ipNum.Add(ipNum, big.NewInt(1)) {
			availableIPs = append(availableIPs, intToIP4(ipNum).String())
		}
	} else if firstIP.To16() != nil {
		// IPv6: no broadcast, so include all except network address
		startIPNum := ip6ToInt(firstIP)
		endIPNum := ip6ToInt(lastIP)
		for ipNum := new(big.Int).Set(startIPNum); ipNum.Cmp(endIPNum) <= 0; ipNum.Add(ipNum, big.NewInt(1)) {
			availableIPs = append(availableIPs, intToIP6(ipNum).String())
		}
	}

	return availableIPs
}

// ip4ToInt converts an IPv4 address to a big.Int.
func ip4ToInt(ip net.IP) *big.Int {
	ip = ip.To4()
	if ip == nil {
		return big.NewInt(0)
	}
	return big.NewInt(0).SetUint64(uint64(binary.BigEndian.Uint32(ip)))
}

// intToIP4 converts a big.Int to an IPv4 address.
func intToIP4(n *big.Int) net.IP {
	maxIPv4 := big.NewInt(int64(math.MaxUint32))
	if n.Cmp(maxIPv4) > 0 {
		n = maxIPv4
	}

	// Get 4-byte representation
	b := n.Bytes()
	if len(b) < 4 {
		// Pad with leading zeros if necessary
		padded := make([]byte, 4)
		copy(padded[4-len(b):], b)
		b = padded
	} else if len(b) > 4 {
		// Trim to last 4 bytes if it's longer (shouldn't happen due to clamp)
		b = b[len(b)-4:]
	}

	return net.IP(b)
}

// ip6ToInt converts an IPv6 address to a big.Int.
func ip6ToInt(ip net.IP) *big.Int {
	ip = ip.To16()
	if ip == nil {
		return big.NewInt(0)
	}
	return big.NewInt(0).SetBytes(ip)
}

// intToIP6 converts a big.Int to an IPv6 address.
func intToIP6(n *big.Int) net.IP {
	b := n.Bytes()
	if len(b) < 16 {
		pad := make([]byte, 0, 16-len(b))
		b = append(pad, b...)
	}
	return net.IP(b)
}

// iPAddressToCIDR converts an IP address with a subnet mask to CIDR format.
// Only supports IPv4 mask notation (e.g., "192.168.1.1/255.255.255.0").
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
// Supports both IPv4 and IPv6.
func networkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP
	mask := network.Mask
	if netIP == nil || mask == nil {
		return nil, nil
	}
	ipLen := len(netIP)
	if ipLen == net.IPv4len {
		netIP = netIP.To4()
	} else if ipLen == net.IPv6len {
		netIP = netIP.To16()
	}
	if netIP == nil {
		return nil, nil
	}
	startIP := make(net.IP, len(netIP))
	copy(startIP, netIP.Mask(mask))
	endIP := make(net.IP, len(startIP))
	for i := range startIP {
		endIP[i] = startIP[i] | ^mask[i]
	}
	return startIP, endIP
}

// ===========================================================================
// =============================   IsLocalhostIP   ===========================
// ===========================================================================

// IsLocalhostIP checks if the given IP address string (ipStr) is bound to any local network interface.
// It returns true if the IP is found on any interface, false otherwise.
// This function parses the input string as an IP address, iterates over all network interfaces on the host,
// and checks if any of the interface addresses match the target IP.
func IsLocalhostIP(ipStr string) bool {
	targetIP := net.ParseIP(ipStr)
	if targetIP == nil {
		// The input string is not a valid IP address.
		return false
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		// Failed to retrieve network interfaces.
		return false
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			// Skip this interface if its addresses cannot be retrieved.
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				// Check if the IP address of this interface matches the target IP.
				if v.IP.Equal(targetIP) {
					return true
				}
			case *net.IPAddr:
				// Check if the IP address of this interface matches the target IP.
				if v.IP.Equal(targetIP) {
					return true
				}
			}
		}
	}
	// The target IP was not found on any local network interface.
	return false
}
