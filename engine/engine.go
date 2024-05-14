package engine

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/docker/go-units"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/option"
	"github.com/xjasonlyu/tun2socks/v2/dialer"
	"github.com/xjasonlyu/tun2socks/v2/engine/mirror"
	"github.com/xjasonlyu/tun2socks/v2/log"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var (
	_engineMutex   sync.Mutex
	_defaultConfig *Config
	_defaultProxy  proxy.Proxy
	_defaultDevice device.Device
	_defaultStack  *stack.Stack
)

func Insert(config *Config) {
	_engineMutex.Lock()
	_defaultConfig = config
	_engineMutex.Unlock()
}

func Start() error {
	_engineMutex.Lock()
	if _defaultConfig == nil {
		return errors.New("empty config")
	}
	for _, f := range []func(*Config) error{
		general,
		netstack,
	} {
		if err := f(_defaultConfig); err != nil {
			return err
		}
	}
	_engineMutex.Unlock()
	return nil
}

func Stop() (err error) {
	_engineMutex.Lock()
	if _defaultDevice != nil {
		err = _defaultDevice.Close()
	}
	if _defaultStack != nil {
		_defaultStack.Close()
		_defaultStack.Wait()
	}
	_engineMutex.Unlock()
	return err
}

func general(k *Config) error {
	level, err := log.ParseLevel(k.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	if k.Interface != "" {
		iface, err := net.InterfaceByName(k.Interface)
		if err != nil {
			return err
		}
		dialer.DefaultInterfaceName.Store(iface.Name)
		dialer.DefaultInterfaceIndex.Store(int32(iface.Index))
		log.Infof("[DIALER] bind to interface: %s", k.Interface)
	}
	if k.UDPTimeout > 0 {
		if k.UDPTimeout < time.Second {
			return errors.New("invalid udp timeout value")
		}
		tunnel.SetUDPTimeout(k.UDPTimeout)
	}
	return nil
}

func netstack(k *Config) (err error) {
	if k.Proxy == "" {
		return errors.New("empty proxy")
	}
	if k.Fd < 0 {
		return errors.New("empty device")
	}
	if _defaultProxy, err = parseProxy(k.Proxy); err != nil {
		return
	}
	proxy.SetDialer(_defaultProxy)
	if _defaultDevice, err = parseDevice(k.Fd); err != nil {
		return
	}
	var multicastGroups []net.IP
	if multicastGroups, err = parseMulticastGroups(k.MulticastGroups); err != nil {
		return err
	}
	var opts []option.Option
	if k.TCPModerateReceiveBuffer {
		opts = append(opts, option.WithTCPModerateReceiveBuffer(true))
	}
	if k.TCPSendBufferSize != "" {
		size, err := units.RAMInBytes(k.TCPSendBufferSize)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTCPSendBufferSize(int(size)))
	}
	if k.TCPReceiveBufferSize != "" {
		size, err := units.RAMInBytes(k.TCPReceiveBufferSize)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTCPReceiveBufferSize(int(size)))
	}
	if _defaultStack, err = core.CreateStack(&core.Config{
		LinkEndpoint:     _defaultDevice,
		TransportHandler: &mirror.Tunnel{},
		MulticastGroups:  multicastGroups,
		Options:          opts,
	}); err != nil {
		return
	}
	log.Infof(
		"[STACK] tun://utun1024 <-> %s://%s",
		_defaultProxy.Proto(), _defaultProxy.Addr(),
	)
	return nil
}
