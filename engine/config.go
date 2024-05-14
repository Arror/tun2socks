package engine

import "time"

type Config struct {
	Fd                       int
	Proxy                    string
	LogLevel                 string
	Interface                string
	TCPModerateReceiveBuffer bool
	TCPSendBufferSize        string
	TCPReceiveBufferSize     string
	MulticastGroups          string
	UDPTimeout               time.Duration
}
