package socks5

type Socks5PasswordAuthServer struct {
	Socks5Server
}

func NewSocks5PasswordAuthServer(ip string, port int, username string, password string, udp bool) *Socks5Server {
	config := Config{
		AuthMethod: MethodPassword,
		PasswordChecker: func(uname, pwd string) bool {
			if username != uname || password != pwd {
				return false
			}
			return true
		},
		UdpPort: UdpRelayClose,
	}
	if udp {
		config.UdpPort = port
	}
	return NewSocks5Server(ip, port, config)
}
