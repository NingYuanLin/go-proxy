package socks5

import (
	"fmt"
	"io"
	"log"
	"net"
)

type TcpRelayServer struct {
	Server *Socks5Server
	Conn   *net.TCPConn
}

func (t *TcpRelayServer) HandleConnection() error {
	// negotiation and sub-negotiation
	err := t.auth()
	if err != nil {
		return err
	}

	// request: establish tcp connect to destination addr or udp associate
	err = t.requestAndForward()
	if err != nil {
		return err
	}

	return nil
}

func (t *TcpRelayServer) auth() error {
	clientMessage, err := NewClientAuthMessage(t.Conn)
	if err != nil {
		return err
	}

	// check if the auth method is supported
	acceptable := false
	for _, method := range clientMessage.Methods {
		if method == t.Server.Config.AuthMethod {
			acceptable = true
			break
		}
	}

	if !acceptable {
		err := WriteServerAuthMessage(t.Conn, MethodNoAcceptable)
		if err != nil {
			return err
		}
		return ErrAuthMethodNotSupport
	}

	err = WriteServerAuthMessage(t.Conn, t.Server.Config.AuthMethod)
	if err != nil {
		return err
	}

	if t.Server.Config.AuthMethod == MethodPassword {
		message, err := NewClientPasswordAuthMessage(t.Conn)
		if err != nil {
			return err
		}
		ok := t.Server.Config.PasswordChecker(message.Username, message.Password)

		if !ok {
			// There is no need to return error because the link will be closed anyway.
			WriteServerPasswordMessage(t.Conn, PasswordAuthFailure)
			return ErrPasswordAuthFailure
		}

		err = WriteServerPasswordMessage(t.Conn, PasswordAuthSuccess)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TcpRelayServer) requestAndForward() error {
	requestMessage, err := NewClientRequestMessage(t.Conn)
	if err != nil {
		return err
	}

	// check if command is supported
	switch requestMessage.Cmd {
	case CmdConnect:
		host := fmt.Sprintf("%s:%d", requestMessage.Address, requestMessage.Port)
		tcpDestConn, err := t.handleTcpRequest(host)
		if err != nil {
			return err
		}
		err = t.forward(tcpDestConn)
		if err != nil {
			return err
		}
	case cmdUdp:
		udpRelayServer, err := t.handleUdpRequest()
		if err != nil {
			return err
		}
		if udpRelayServer != nil {
			err := udpRelayServer.HandleConnection()
			if err != nil {
				log.Println(err)
			}
		}
	default:
		WriteRequestFailureReply(t.Conn, ReplyCommandNotSupported)
		return ErrCommandNotSupport
	}
	return nil
}

// handleTcpRequest
// host: similar to "x.x.x.x:x"
func (t *TcpRelayServer) handleTcpRequest(host string) (io.ReadWriteCloser, error) {
	// access destination address
	destConn, err := net.DialTimeout("tcp", host, t.Server.Config.Timeout)
	if err != nil {
		WriteRequestFailureReply(t.Conn, ReplyNetworkUnreachable)
		return nil, err
	}

	// send success reply
	tcpAddr := destConn.LocalAddr().(*net.TCPAddr)
	err = WriteRequestSuccessReply(t.Conn, tcpAddr.IP, uint16(tcpAddr.Port))
	if err != nil {
		WriteRequestFailureReply(t.Conn, ReplyServerFailure)
		return nil, err
	}

	return destConn, nil
}

// When udp relay is not opened, it will return nil and ErrCommandNotSupport.
// When t.Server.Config.UdpPort is UdpRelayRandomPort, it will reply a successful udp associate request and return a new *UdpRelayServer and nil.
// When use concrete udp relay port, it will reply a successful udp associate request and return nil and nil.
func (t *TcpRelayServer) handleUdpRequest() (*UdpRelayServer, error) {
	if t.Server.Config.UdpPort == UdpRelayClose {
		WriteRequestFailureReply(t.Conn, ReplyConnectionNotAllowed)
		return nil, ErrCommandNotSupport
	} else if t.Server.Config.UdpPort == UdpRelayRandomPort {
		conn, err := NewUdpConn(":0")
		if err != nil {
			return nil, err
		}

		udpRelayServerIp := t.Server.Config.UdpRelayServerIp
		if udpRelayServerIp == nil {
			// use ip of tcp connection
			udpRelayServerIp = t.Conn.LocalAddr().(*net.TCPAddr).IP
		}

		port := conn.LocalAddr().(*net.UDPAddr).Port
		err = WriteRequestSuccessReply(t.Conn, udpRelayServerIp, uint16(port))
		if err != nil {
			return nil, err
		}

		udpRelayServer := NewUdpRelayServer(t.Server, conn, t.Conn)
		return udpRelayServer, nil
	} else {
		udpRelayServerIp := t.Server.Config.UdpRelayServerIp
		if udpRelayServerIp == nil {
			udpRelayServerIp = t.Conn.LocalAddr().(*net.TCPAddr).IP
		}

		err := WriteRequestSuccessReply(t.Conn, udpRelayServerIp, uint16(t.Server.Config.UdpPort))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (t *TcpRelayServer) forward(destConn io.ReadWriteCloser) error {
	defer destConn.Close()
	go io.Copy(destConn, t.Conn)
	_, err := io.Copy(t.Conn, destConn)
	return err
}
