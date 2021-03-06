package main

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
)

// Interface provides the information for a device interface
type Interface struct {
	HardwareAddr string
	IPv4Addr     string
	IPv4Mask     string
	IPv4Network  string
	IPv6Addr     string
	MTU          int
	Name         string
}

// New instantiates an Interface object for the passed in net.Interface type
// representing a device interface
func New(iface net.Interface) (i *Interface, err error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return i, err
	}

	// we can't rely on the order of the addresses in the addrs array
	ipv4, ipv6 := extractAddrs(addrs)

	// if we have an IPv6 only interface
	if len(ipv4) == 0 {
		return &Interface{
			Name:     iface.Name,
			IPv6Addr: ipv6,
		}, nil
	}

	// get IPv4 address & dotted decimal mask
	ipv4Split := strings.Split(ipv4, "/")
	ipv4addr := ipv4Split[0]
	ipv4Cidr := ipv4Split[1]
	ipv4Mask, err := toDottedDec(ipv4Cidr)
	if err != nil {
		return i, err
	}

	// get IPv4 network
	_, ipnet, err := net.ParseCIDR(ipv4)
	if err != nil {
		return i, err
	}

	return &Interface{
		HardwareAddr: iface.HardwareAddr.String(),
		IPv4Addr:     ipv4addr,
		IPv4Mask:     ipv4Mask,
		IPv4Network:  ipnet.String(),
		IPv6Addr:     ipv6,
		MTU:          iface.MTU,
		Name:         iface.Name,
	}, nil
}

func extractAddrs(addrs []net.Addr) (ipv4 string, ipv6 string) {
	for _, addr := range addrs {
		if a := addr.String(); strings.Contains(a, ":") {
			ipv6 = a
		}
		if a := addr.String(); strings.Contains(a, ".") {
			ipv4 = a
		}
	}
	return
}

func toDottedDec(cidr string) (s string, err error) {
	maskBits := []string{"", "128", "192", "224", "240", "248", "252", "254", "255"}
	n, err := strconv.Atoi(cidr)
	if err != nil {
		return s, err
	}

	if n > 32 || n < 0 {
		return s, fmt.Errorf("Not a valid network mask: %s", cidr)
	}

	allOnes := n / 8
	someOnes := n % 8
	mask := make([]string, 4)

	for i := 0; i < allOnes; i++ {
		mask[i] = "255"
	}

	if maskBits[someOnes] != "" {
		mask[allOnes] = maskBits[someOnes]
	}

	for i, octet := range mask {
		if octet == "" {
			mask[i] = "0"
		}
	}

	return strings.Join(mask, "."), nil
}

func (iface *Interface) String() string {
	ifaceString := "%-10s  %-15s  %-15s  %-18s  %4d  %17s  %s"
	if runtime.GOOS == "windows" {
		ifaceString = "%-35s  %-15s  %-15s  %-18s  %4d  %17s  %s"
	}
	return fmt.Sprintf(
		ifaceString,
		iface.Name,
		iface.IPv4Addr,
		iface.IPv4Mask,
		iface.IPv4Network,
		iface.MTU,
		iface.HardwareAddr,
		iface.IPv6Addr,
	)
}
