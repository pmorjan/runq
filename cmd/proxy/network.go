package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gotoz/runq/internal/util"
	"github.com/gotoz/runq/pkg/vm"
	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"
)

func setupNetwork() ([]vm.Network, error) {
	var networks []vm.Network
	var idx int

	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("netlink.LinkList() failed: %w", err)
	}

	for _, link := range links {
		switch link.Type() {
		case "veth", "macvlan":
		default:
			continue
		}

		linkAttrs := link.Attrs()

		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return nil, fmt.Errorf("netlink.AddrList() failed: %w", err)
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("no ip found on %s", linkAttrs.Name)
		}

		var gateway net.IP
		routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
		if err != nil {
			return nil, fmt.Errorf("netlink.RouteList() failed: %w", err)
		}
		for _, route := range routes {
			if route.Gw != nil {
				gateway = route.Gw
				break
			}
		}

		for _, a := range addrs {
			err = netlink.AddrDel(link, &a)
			if err != nil {
				return nil, fmt.Errorf("netlink.AddrDel() failed: %w", err)
			}
		}

		if err = netlink.LinkSetDown(link); err != nil {
			return nil, fmt.Errorf("netlink.LinkSetDown() failed: %w", err)
		}

		// create macvtap interface
		mvtAttrs := netlink.NewLinkAttrs()
		mvtAttrs.Name = fmt.Sprintf("tap%d", idx)
		mvtAttrs.ParentIndex = linkAttrs.Index
		mvt := &netlink.Macvtap{
			Macvlan: netlink.Macvlan{
				LinkAttrs: mvtAttrs,
				Mode:      netlink.MACVLAN_MODE_BRIDGE,
			},
		}

		if err := netlink.LinkAdd(mvt); err != nil {
			return nil, fmt.Errorf("netlink.LinkAdd macvtap interface failed: %w", err)
		}

		macvtap, err := netlink.LinkByName(mvtAttrs.Name)
		if err != nil {
			return nil, fmt.Errorf("netlink.LinkByName failed, name:%s : %w", mvtAttrs.Name, err)
		}
		mvtAttrs = *macvtap.Attrs()

		if err := netlink.LinkSetUp(macvtap); err != nil {
			return nil, fmt.Errorf("netlink.LinkSetUp macvtap interface failed, name:%s : %w", mvtAttrs.Name, err)
		}

		if err := netlink.LinkSetUp(link); err != nil {
			return nil, fmt.Errorf("netlink.LinkSetUp  interface failed, name:%s : %w", linkAttrs.Name, err)
		}

		tapDevice, err := createTapDevice(mvtAttrs.Name, mvtAttrs.Index)
		if err != nil {
			return nil, err
		}

		networks = append(networks, vm.Network{
			Name:       linkAttrs.Name,
			MacAddress: mvtAttrs.HardwareAddr.String(),
			MTU:        linkAttrs.MTU,
			Addrs:      addrs,
			Gateway:    gateway,
			TapDevice:  tapDevice,
		})

		idx++
	}

	return networks, nil
}

func createTapDevice(name string, index int) (string, error) {
	syspath := fmt.Sprintf("/sys/devices/virtual/net/%s/tap%d/dev", name, index)
	major, minor, err := util.MajorMinor(syspath)
	if err != nil {
		return "", err
	}
	devpath := "/dev/" + name
	if err := util.Mknod(devpath, "c", 0600, major, minor); err != nil {
		return "", err
	}
	return devpath, nil
}

func writeResolvConf(dns vm.DNS) error {
	if dns.Preserve {
		return nil
	}
	const file = "/etc/resolv.conf"

	str := "# Generated by RunQ\n"
	if dns.Options != "" {
		str += fmt.Sprintf("options %s\n", dns.Options)
	}
	if dns.Search != "" {
		str += fmt.Sprintf("search %s\n", dns.Search)
	}

	// A given dns server can be a regular IP or a hostname.
	// If it is a hostname, it is asumed that it's a DNS proxy container.
	var err error
	for _, ns := range dns.Server {
		ip := net.ParseIP(ns)
		if ip == nil {
			ip, err = proxyIP(ns)
			if err != nil {
				return err
			}
		}
		str += fmt.Sprintf("nameserver %s\n", ip)
	}

	if err := unix.Unmount(file, unix.MNT_DETACH); err != nil {
		if err != unix.EINVAL {
			return err
		}
		// file was not a mount point. Ensure /etc exists.
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(file, []byte(str), 0444); err != nil {
		return err
	}
	return os.Chmod(file, 0444)
}

// proxyIP tries to find the IP address of a given hostname via the local resolver.
// The proxyIP is valid only if it is part of a local network.
func proxyIP(host string) (net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1000)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		valid, err := proxyIPisValid(ip.IP)
		if err != nil {
			return nil, err
		}
		if valid {
			return ip.IP, nil
		}
	}
	return nil, fmt.Errorf("invalid nameserver %q", host)
}

// proxyIPisValid checks if the given ip exists in one of the local networks.
func proxyIPisValid(ip net.IP) (bool, error) {
	if ip.To4() == nil {
		return false, nil
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}
	for _, a := range addrs {
		ifip, net, err := net.ParseCIDR(a.String())
		if err != nil {
			return false, err
		}
		if ifip.To4() == nil {
			continue
		}
		if ifip.IsLoopback() {
			continue
		}
		if net.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}
