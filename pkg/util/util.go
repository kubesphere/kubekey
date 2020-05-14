package util

import (
	"encoding/binary"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	VERSION = "KubeKey Version v0.0.1\nKubernetes Version v1.17.4"
)

func InitLogger(verbose bool) *log.Logger {
	logger := log.New()
	logger.Formatter = &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 MST",
	}

	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	return logger
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(tmpl *template.Template, variables map[string]interface{}) (string, error) {

	var buf strings.Builder
	//buf.WriteString(`set -xeu pipefail`)
	//buf.WriteString("\n\n")
	//buf.WriteString(`export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"`)
	//buf.WriteString("\n\n")

	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "Failed to render cmd or script template")
	}
	return buf.String(), nil
}

func ParseIp(ip string) []string {
	var availableIPs []string
	// if ip is "1.1.1.1/",trim /
	ip = strings.TrimRight(ip, "/")
	if strings.Contains(ip, "/") == true {
		if strings.Contains(ip, "/32") == true {
			aip := strings.Replace(ip, "/32", "", -1)
			availableIPs = append(availableIPs, aip)
		} else {
			availableIPs = GetAvailableIP(ip)
		}
	} else if strings.Contains(ip, "-") == true {
		ipRange := strings.SplitN(ip, "-", 2)
		availableIPs = GetAvailableIPRange(ipRange[0], ipRange[1])
	} else {
		availableIPs = append(availableIPs, ip)
	}
	return availableIPs
}

func GetAvailableIPRange(ipStart, ipEnd string) []string {
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
		newNum = newNum + pos
	}
	return availableIPs
}

func GetAvailableIP(ipAndMask string) []string {
	var availableIPs []string

	ipAndMask = strings.TrimSpace(ipAndMask)
	ipAndMask = IPAddressToCIDR(ipAndMask)
	_, ipnet, _ := net.ParseCIDR(ipAndMask)

	firstIP, _ := networkRange(ipnet)
	ipNum := ipToInt(firstIP)
	size := networkSize(ipnet.Mask)
	pos := int32(1)
	max := size - 2 // -1 for the broadcast address, -1 for the gateway address

	var newNum int32
	for attempt := int32(0); attempt < max; attempt++ {
		newNum = ipNum + pos
		pos = pos%max + 1
		availableIPs = append(availableIPs, intToIP(newNum).String())
	}
	return availableIPs
}

func IPAddressToCIDR(ipAdress string) string {
	if strings.Contains(ipAdress, "/") == true {
		ipAndMask := strings.Split(ipAdress, "/")
		ip := ipAndMask[0]
		mask := ipAndMask[1]
		if strings.Contains(mask, ".") == true {
			mask = IPMaskStringToCIDR(mask)
		}
		return ip + "/" + mask
	} else {
		return ipAdress
	}
}

func IPMaskStringToCIDR(netmask string) string {
	netmaskList := strings.Split(netmask, ".")
	var mint []int
	for _, v := range netmaskList {
		strv, _ := strconv.Atoi(v)
		mint = append(mint, strv)
	}
	myIPMask := net.IPv4Mask(byte(mint[0]), byte(mint[1]), byte(mint[2]), byte(mint[3]))
	ones, _ := myIPMask.Size()
	return strconv.Itoa(ones)
}

func IPMaskCIDRToString(one string) string {
	oneInt, _ := strconv.Atoi(one)
	mIPmask := net.CIDRMask(oneInt, 32)
	var maskstring []string
	for _, v := range mIPmask {
		maskstring = append(maskstring, strconv.Itoa(int(v)))
	}
	return strings.Join(maskstring, ".")
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

func ipToInt(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

func intToIP(n int32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && ipnet.IP.IsGlobalUnicast() {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("valid local IP not found!")
}

func LocalIP() string {
	localIp, err := GetLocalIP()
	if err != nil {
		log.Fatalf("Failed to get Local IP: %v", err)
	}
	return localIp
}
