package socks5

import (
	"encoding/binary"
	"io"
	"net"
)

const (
	Ipv4Length = 4
	Ipv6Length = 16
	PortLength = 2
)

type ReplyType = byte

const (
	ReplySuccess ReplyType = iota
	ReplyServerFailure
	ReplyConnectionNotAllowed
	ReplyNetworkUnreachable
	ReplyHostUnreachable
	ReplyConnectionRefused
	ReplyTTLExpired
	ReplyCommandNotSupported
	ReplyAddressTypeNotSupported
)

type ClientRequestMessage struct {
	Cmd         Command
	Address     string
	Port        uint16
	AddressType AddressType
}

type Command = byte

const (
	CmdConnect Command = 0x01
	CmdBind    Command = 0x02
	cmdUdp     Command = 0x03
)

func NewClientRequestMessage(conn io.ReadWriter) (*ClientRequestMessage, error) {
	// Read version, command, reserved, address type
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}

	version, command, reversed, addressType := buf[0], buf[1], buf[2], buf[3]

	// check if field is valid
	if version != Socks5Version {
		return nil, ErrVersionNotSupport
	}
	if command != CmdConnect && command != CmdBind && command != cmdUdp {
		return nil, ErrCommandNotSupport
	}
	if reversed != ReversedField {
		return nil, ErrInvalidReversedField
	}
	if addressType != AddressTypeIpv4 && addressType != AddressTypeIpv6 && addressType != AddressTypeDomain {
		return nil, ErrAddressTypeNotSupport
	}

	message := ClientRequestMessage{
		Cmd:         command,
		AddressType: addressType,
	}
	// read address
	switch addressType {
	case AddressTypeIpv6:
		buf = make([]byte, Ipv6Length)
		fallthrough
	case AddressTypeIpv4:
		if _, err = io.ReadFull(conn, buf); err != nil {
			return nil, err
		}
		ip := net.IP(buf)
		message.Address = ip.String()
	case AddressTypeDomain:
		if _, err = io.ReadFull(conn, buf[:1]); err != nil {
			return nil, err
		}
		domainLength := buf[0]
		if domainLength > Ipv4Length {
			buf = make([]byte, domainLength)
		}
		_, err = io.ReadFull(conn, buf[:domainLength])
		if err != nil {
			return nil, err
		}
		message.Address = string(buf[:domainLength])
	}

	// read port
	if _, err = io.ReadFull(conn, buf[:PortLength]); err != nil {
		return nil, err
	}
	message.Port = (uint16(buf[0]) << 8) + uint16(buf[1])

	return &message, nil
}

func WriteRequestSuccessReply(conn io.Writer, ip net.IP, port uint16) error {
	// There must create connBuf to avoid "Packet Fragmentation"
	// When Packet Fragmentation occurs, error may be happened in ss-tap.
	// 22: response length when there is ipv6
	connBuf := make([]byte, 0, 22)
	connBuf = append(connBuf, Socks5Version)
	connBuf = append(connBuf, ReplySuccess)
	connBuf = append(connBuf, ReversedField)

	// write AddressType and ip
	if ip4 := ip.To4(); ip4 != nil {
		connBuf = append(connBuf, AddressTypeIpv4)
		connBuf = append(connBuf, ip4...)
	} else {
		// ipv6
		connBuf = append(connBuf, AddressTypeIpv6)
		connBuf = append(connBuf, ip...)
	}

	// write port
	buf := make([]byte, 0, 2)
	buf = binary.BigEndian.AppendUint16(buf, port)
	connBuf = append(connBuf, buf...)

	_, err := conn.Write(connBuf)
	if err != nil {
		return err
	}

	return nil
}

func WriteRequestFailureReply(conn io.Writer, replyType ReplyType) error {
	_, err := conn.Write([]byte{Socks5Version, replyType, ReversedField, AddressTypeIpv4, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return err
	}

	return nil
}
