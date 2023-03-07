package socks5

func NewSocks5NoAuthServer(ip string, port int, udp bool) *Socks5Server {
	config := Config{
		AuthMethod: MethodNoAuth,
		UdpPort:    UdpRelayClose,
	}
	if udp {
		config.UdpPort = port
	}
	return NewSocks5Server(ip, port, config)
}
