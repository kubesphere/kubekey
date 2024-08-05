package internal

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
)

// parseIp parse cidr to actual ip slice. or parse the ip range string (format xxx-xxx) to actual ip slice,
func parseIp(ip string) []string {
	var availableIPs []string
	// if ip is "1.1.1.1/",trim /
	ip = strings.TrimRight(ip, "/")
	if strings.Contains(ip, "/") {
		if strings.Contains(ip, "/32") {
			aip := strings.Replace(ip, "/32", "", -1)
			availableIPs = append(availableIPs, aip)
		} else {
			availableIPs = getAvailableIP(ip)
		}
	} else if strings.Contains(ip, "-") {
		ipRange := strings.SplitN(ip, "-", 2)
		availableIPs = getAvailableIPRange(ipRange[0], ipRange[1])
	} else {
		availableIPs = append(availableIPs, ip)
	}
	return availableIPs
}

func getAvailableIPRange(ipStart, ipEnd string) []string {
	var availableIPs []string

	firstIP := net.ParseIP(ipStart)
	endIP := net.ParseIP(ipEnd)
	if firstIP.To4() == nil || endIP.To4() == nil {
		return availableIPs
	}
	firstIPNum := ipToInt(firstIP.To4())
	EndIPNum := ipToInt(endIP.To4())
	pos := int32(1)

	newNum := firstIPNum

	for newNum <= EndIPNum {
		availableIPs = append(availableIPs, intToIP(newNum).String())
		newNum += pos
	}
	return availableIPs
}

func getAvailableIP(ipAndMask string) []string {
	var availableIPs []string

	ipAndMask = strings.TrimSpace(ipAndMask)
	ipAndMask = iPAddressToCIDR(ipAndMask)
	_, ipnet, _ := net.ParseCIDR(ipAndMask)

	firstIP, _ := networkRange(ipnet)
	ipNum := ipToInt(firstIP)
	size := networkSize(ipnet.Mask)
	pos := int32(1)
	m := size - 2 // -1 for the broadcast address, -1 for the gateway address

	var newNum int32
	for attempt := int32(0); attempt < m; attempt++ {
		newNum = ipNum + pos
		pos = pos%m + 1
		availableIPs = append(availableIPs, intToIP(newNum).String())
	}
	return availableIPs
}

func ipToInt(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

func intToIP(n int32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
}

func iPAddressToCIDR(ipAddress string) string {
	if strings.Contains(ipAddress, "/") {
		ipAndMask := strings.Split(ipAddress, "/")
		ip := ipAndMask[0]
		mask := ipAndMask[1]
		if strings.Contains(mask, ".") {
			mask = iPMaskStringToCIDR(mask)
		}
		return ip + "/" + mask
	} else {
		return ipAddress
	}
}

func iPMaskStringToCIDR(netmask string) string {
	netmaskList := strings.Split(netmask, ".")
	var mint = make([]int, len(netmaskList))
	for i, v := range netmaskList {
		mint[i], _ = strconv.Atoi(v)
	}
	myIPMask := net.IPv4Mask(byte(mint[0]), byte(mint[1]), byte(mint[2]), byte(mint[3]))
	ones, _ := myIPMask.Size()
	return strconv.Itoa(ones)
}

func networkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP.To4()
	firstIP := netIP.Mask(network.Mask)
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
}

func networkSize(mask net.IPMask) int32 {
	m := net.IPv4Mask(0, 0, 0, 0)
	for i := 0; i < net.IPv4len; i++ {
		m[i] = ^mask[i]
	}
	return int32(binary.BigEndian.Uint32(m)) + 1
}
