package socks5

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	ErrVersionNotSupport     = errors.New("protocol version not supported")
	ErrCommandNotSupport     = errors.New("request command not supported")
	ErrInvalidReversedField  = errors.New("reversed field is invalid")
	ErrAddressTypeNotSupport = errors.New("address type not supported")

	ErrAuthMethodNotSupport = errors.New("authentication method not supported")

	ErrMethodVersionNotSupported = errors.New("sub-negotiation method not supported")

	ErrPasswordCheckerNotSet = errors.New("passwordChecker not set")

	ErrUnknownAddr   = errors.New("address not supported")
	ErrUdpPortListen = errors.New("udp port open failed")
)

const (
	Socks5Version = 0x05
	ReversedField = 0x00
)

type AddressType = byte

const (
	AddressTypeIpv4   AddressType = 0x01
	AddressTypeDomain AddressType = 0x03
	AddressTypeIpv6   AddressType = 0x04
)

type UdpRelayPort = int

const (
	UdpRelayClose      UdpRelayPort = -1 // close udp relay
	UdpRelayRandomPort UdpRelayPort = 0  //	random port
)

type PasswordCheckerFunc func(username, password string) bool

type Config struct {
	AuthMethod      AuthMethod
	Timeout         time.Duration       // The timeout of tcp dial.
	PasswordChecker PasswordCheckerFunc // only for username/password authentication

	// The ip of udp relay server
	// By default, when UdpRelayServerIp == nil, the system will detect it automatically.
	// If your relay server is under NAT, you may need to set it manually.
	UdpRelayServerIp net.IP
	// -1 means close udp relay. 0 means use random port. Concrete number, such as 1080, means use fixed port(not safe).
	// Note: when use random port, you should open your firewall to allow all udp access from any address.
	UdpPort UdpRelayPort
	// The lifetime of udp exchange socket.
	UdpConnLifetime time.Duration
}

type Server interface {
	Run() error
}

// Socks5Server
// Do not create it directly. Please use concrete constructor function
// such as NewSocks5NoAuthServer, NewSocks5PasswordAuthServer.
type Socks5Server struct {
	// The listened ip
	// It can be "0.0.0.0" to allow all ipv4 request or "127.0.0.1" to only allow localhost request.
	Ip string
	// The tcp listen port
	Port int
	//UdpRelayInfo *UdpRelayInfo
	Config Config
}

func NewSocks5Server(ip string, port int, config Config) *Socks5Server {
	server := &Socks5Server{
		Ip:     ip,
		Port:   port,
		Config: config,
	}
	return server
}

func (s *Socks5Server) SetTimeout(timeout time.Duration) {
	s.Config.Timeout = timeout
}

func (s *Socks5Server) init() {
	if s.Config.UdpConnLifetime == 0 {
		s.Config.UdpConnLifetime = time.Second * 60
	}
	if s.Config.Timeout == 0 {
		s.Config.Timeout = 3
	}
}

func (s *Socks5Server) Run() error {
	s.init()

	addr := fmt.Sprintf("%s:%d", s.Ip, s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// concrete udp listen port
	if s.Config.UdpPort != UdpRelayClose && s.Config.UdpPort != UdpRelayRandomPort {
		udpConn, err := NewUdpConn(fmt.Sprintf(":%d", s.Config.UdpPort))
		if err != nil {
			return err
		}
		udpRelayServer := NewUdpRelayServer(s, udpConn, nil)
		go func() {
			// udp connection should work through all lifetime of the program
			for {
				err := udpRelayServer.HandleConnection()
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}

	// accept tcp connection
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection failure from :%s Err message:%s", conn.RemoteAddr(), err.Error())
		}
		go func() {
			defer conn.Close()
			tcpRelayServer := TcpRelayServer{
				Server: s,
				Conn:   conn.(*net.TCPConn),
			}
			err := tcpRelayServer.HandleConnection()
			if err != nil {
				log.Printf("Handle connection failure from :%s Err message:%s", conn.RemoteAddr(), err.Error())
			}
		}()
	}
}
