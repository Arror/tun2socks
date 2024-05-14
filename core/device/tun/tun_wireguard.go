package tun

import (
	"fmt"
	"os"
	"sync"

	"golang.zx2c4.com/wireguard/tun"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
)

type TUN struct {
	*iobased.Endpoint

	nt     *tun.NativeTun
	offset int

	rSizes []int
	rBuffs [][]byte
	wBuffs [][]byte
	rMutex sync.Mutex
	wMutex sync.Mutex
}

func Open(fd int) (_ device.Device, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("open tun: %v", r)
		}
	}()
	t := &TUN{
		offset: 4,
		rSizes: make([]int, 1),
		rBuffs: make([][]byte, 1),
		wBuffs: make([][]byte, 1),
	}
	nt, err := tun.CreateTUNFromFile(os.NewFile(uintptr(fd), "utun"), 0)
	if err != nil {
		return nil, fmt.Errorf("create tun: %w", err)
	}
	t.nt = nt.(*tun.NativeTun)
	mtu, err := nt.MTU()
	if err != nil {
		return nil, fmt.Errorf("get mtu: %w", err)
	}
	ep, err := iobased.New(t, uint32(mtu), 4)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	t.Endpoint = ep
	return t, nil
}

func (t *TUN) Read(packet []byte) (int, error) {
	t.rMutex.Lock()
	defer t.rMutex.Unlock()
	t.rBuffs[0] = packet
	_, err := t.nt.Read(t.rBuffs, t.rSizes, t.offset)
	return t.rSizes[0], err
}

func (t *TUN) Write(packet []byte) (int, error) {
	t.wMutex.Lock()
	defer t.wMutex.Unlock()
	t.wBuffs[0] = packet
	return t.nt.Write(t.wBuffs, t.offset)
}

func (t *TUN) Close() error {
	defer t.Endpoint.Close()
	return t.nt.Close()
}
