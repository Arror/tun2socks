package engine

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/tun"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
)

func parseDevice(fd int) (device.Device, error) {
	return tun.Open(fd)
}

func parseProxy(s string) (proxy.Proxy, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", proto.Socks5 /* default protocol */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	protocol := strings.ToLower(u.Scheme)

	switch protocol {
	case proto.Direct.String():
		return proxy.NewDirect(), nil
	case proto.Reject.String():
		return proxy.NewReject(), nil
	case proto.Socks5.String():
		return parseSocks5(u)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func parseSocks5(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()

	// Socks5 over UDS
	if address == "" {
		address = u.Path
	}
	return proxy.NewSocks5(address, username, password)
}

func parseMulticastGroups(s string) (multicastGroups []net.IP, _ error) {
	ipStrings := strings.Split(s, ",")
	for _, ipString := range ipStrings {
		if strings.TrimSpace(ipString) == "" {
			continue
		}
		ip := net.ParseIP(ipString)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP format: %s", ipString)
		}
		if !ip.IsMulticast() {
			return nil, fmt.Errorf("invalid multicast IP address: %s", ipString)
		}
		multicastGroups = append(multicastGroups, ip)
	}
	return
}
