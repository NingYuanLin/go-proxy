package socks5

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const MaxUdpBufLength = 65507

var ErrOpenUdpConnection = errors.New("open udp port failed")

type UdpExchange struct {
	ExpiredTime    time.Time
	DConn          *net.UDPConn // connection with destination
	UdpRelayServer *UdpRelayServer
	ClientAddr     *net.UDPAddr
	Closed         chan struct{} // prepare for closing
	ClosedOk       chan struct{} // have closed
}

func NewUdpExchange(conn *net.UDPConn, lifetime time.Duration, udpRelayServer *UdpRelayServer, clientAddr *net.UDPAddr) *UdpExchange {
	udpExchange := &UdpExchange{}
	udpExchange.ExpiredTime = time.Now().Add(lifetime)
	udpExchange.DConn = conn
	udpExchange.UdpRelayServer = udpRelayServer
	udpExchange.ClientAddr = clientAddr
	udpExchange.Closed = make(chan struct{})
	udpExchange.ClosedOk = make(chan struct{})
	return udpExchange
}

func (u *UdpExchange) IsExpired() bool {
	if time.Now().Unix() > u.ExpiredTime.Unix() {
		return true
	}
	return false
}

func (u *UdpExchange) Refresh(lifetime time.Duration) {
	u.ExpiredTime = time.Now().Add(lifetime)
}

func (u *UdpExchange) Close() {
	u.Closed <- struct{}{}
	<-u.ClosedOk // waiting for all work to be completed
}

func (u *UdpExchange) Handle() error {
	buf := make([]byte, MaxUdpBufLength)
	defer func() {
		u.DConn.Close()
	}()

	for {
		select {
		case <-u.Closed:
			u.ClosedOk <- struct{}{}
			return nil
		default:
			err := u.DConn.SetReadDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				return err
			}

			n, addr, err := u.DConn.ReadFrom(buf)
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else if err != nil {
				return err
			}

			u.Refresh(u.UdpRelayServer.Server.Config.UdpConnLifetime)

			toClientBytes, err := NewUdpServerForwardBytes(addr, buf[:n])
			if err != nil {
				return err
			}
			_, err = u.UdpRelayServer.Conn.WriteToUDP(toClientBytes, u.ClientAddr)
			//_, err = u.UdpRelayServer.Conn.WriteTo(toClientBytes, u.ClientAddr)
			if err != nil {
				return err
			}
		}
	}
}

type UdpRelayServer struct {
	Server            *Socks5Server
	Conn              *net.UDPConn            // connection with client
	TcpConn           *net.TCPConn            // may be nil
	UdpExchanges      map[string]*UdpExchange // host to connection with destination
	UdpExchangesMutex sync.Mutex
}

// NewUdpRelayServer is defined to Create a new UdpRelayServer.
// TcpConn present the tcp connection during auth and request
// When tcpConn closed, the udp connection will be closed.
func NewUdpRelayServer(server *Socks5Server, conn *net.UDPConn, tcpConn *net.TCPConn) *UdpRelayServer {
	udpRelayServer := &UdpRelayServer{}
	udpRelayServer.Server = server
	udpRelayServer.Conn = conn
	udpRelayServer.TcpConn = tcpConn

	udpRelayServer.UdpExchanges = make(map[string]*UdpExchange)

	return udpRelayServer
}

func (u *UdpRelayServer) Close() error {
	// step1. close the connection to the destination
	u.UdpExchangesMutex.Lock()
	for _, udpExchange := range u.UdpExchanges {
		udpExchange.Close()
	}
	u.UdpExchangesMutex.Unlock()

	// step2. close the udp connection to the client
	err := u.Conn.Close()

	// step3. close the tcp connection to the client
	if u.TcpConn != nil {
		err = u.TcpConn.Close()
	}

	return err
}

func (u *UdpRelayServer) HandleConnection() error {
	defer u.Close()
	tcpDone := make(chan struct{})
	handleDone := make(chan struct{})
	defer close(handleDone)

	if u.TcpConn != nil {
		go func() {
			defer func() {
				tcpDone <- struct{}{}
			}()
			buf := make([]byte, 1)
			for {
				_, err := u.TcpConn.Read(buf)
				if err != nil {
					return
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				u.UdpExchangesMutex.Lock()
				for host, udpExchange := range u.UdpExchanges {
					if udpExchange.IsExpired() {
						udpExchange.Close()
						delete(u.UdpExchanges, host)
					}
				}
				u.UdpExchangesMutex.Unlock()
			case <-handleDone:
				return
			}
		}
	}()

	buf := make([]byte, MaxUdpBufLength)
	for {
		select {
		case <-tcpDone:
			return nil
		default:
			err := u.Conn.SetReadDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				return err
			}
			n, addr, err := u.Conn.ReadFromUDP(buf)
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else if err != nil {
				return err
			}

			host := fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
			//fmt.Println(host)
			u.UdpExchangesMutex.Lock()
			udpExchange, ok := u.UdpExchanges[host]
			if !ok {
				// create a new udp conn and start to handle
				dConn, err := NewUdpConn(":0")
				if err != nil {
					return ErrOpenUdpConnection
				}
				//fmt.Println(u.Server.Config.UdpConnLifetime)
				udpExchange = NewUdpExchange(dConn, u.Server.Config.UdpConnLifetime, u, addr)
				u.UdpExchanges[host] = udpExchange
				go func() {
					host := host
					err := udpExchange.Handle()
					if err != nil {
						u.UdpExchangesMutex.Lock()
						if _, ok := u.UdpExchanges[host]; ok {
							delete(u.UdpExchanges, host)
						}
						u.UdpExchangesMutex.Unlock()
						log.Println("udp exchange error: ", err)
					}
				}()
			}
			u.UdpExchangesMutex.Unlock()

			udpClientForwardMessage, err := NewUdpClientForwardMessage(buf[:n])
			if err != nil {
				return err
			}

			// resolve addr
			udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", udpClientForwardMessage.Address, udpClientForwardMessage.Port))
			if err != nil {
				return err
			}

			udpExchange.Refresh(u.Server.Config.UdpConnLifetime)
			_, err = udpExchange.DConn.WriteTo(udpClientForwardMessage.Data, udpAddr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NewUdpConn
// host == ":0" means use a random port
func NewUdpConn(host string) (*net.UDPConn, error) {
	//fmt.Println(host)
	udpAddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return udpConn, nil
}
