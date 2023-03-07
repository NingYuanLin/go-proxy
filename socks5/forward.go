package socks5

import (
	"encoding/binary"
	"errors"
	"net"
	"reflect"
	"strconv"
)

var (
	ErrUdpForwardVersionNotSupported = errors.New("udp forward version not supported")
	ErrUdpReassembleNotSupported     = errors.New("udp frame reassemble not supported")
)

var UdpForwardVersion = []byte{0, 0}

type UdpClientForwardMessage struct {
	FPAG        byte // 0x00 means complete. otherwise means
	AddressType AddressType
	Address     string
	Port        uint16
	Data        []byte
}

func NewUdpClientForwardMessage(bytes []byte) (*UdpClientForwardMessage, error) {
	udpClientForwardMessage := &UdpClientForwardMessage{}
	version := bytes[:2]
	if !reflect.DeepEqual(version, UdpForwardVersion) {
		return nil, ErrUdpForwardVersionNotSupported
	}

	FPAG := bytes[2] // Current fragment number
	if FPAG != 0x00 {
		return nil, ErrUdpReassembleNotSupported
	}

	addressType := bytes[3]
	if addressType != AddressTypeIpv4 && addressType != AddressTypeIpv6 && addressType != AddressTypeDomain {
		return nil, ErrAddressTypeNotSupport
	}
	udpClientForwardMessage.AddressType = addressType

	startAddrIndex := 4
	var endAddrIndex int
	switch addressType {
	case AddressTypeIpv4:
		endAddrIndex = startAddrIndex + Ipv4Length
	case AddressTypeIpv6:
		endAddrIndex = startAddrIndex + Ipv6Length
	case AddressTypeDomain:
		domainLength := bytes[startAddrIndex]
		startAddrIndex += 1
		endAddrIndex = startAddrIndex + int(domainLength)
	}
	addr := bytes[startAddrIndex:endAddrIndex]
	switch addressType {
	case AddressTypeIpv4:
		fallthrough
	case AddressTypeIpv6:
		udpClientForwardMessage.Address = net.IP(addr).String()
	case AddressTypeDomain:
		udpClientForwardMessage.Address = string(addr)
	}

	port := binary.BigEndian.Uint16(bytes[endAddrIndex : endAddrIndex+2])
	udpClientForwardMessage.Port = port

	data := bytes[endAddrIndex+2:]
	udpClientForwardMessage.Data = data

	return udpClientForwardMessage, nil

}

func NewUdpServerForwardBytes(addr net.Addr, data []byte) ([]byte, error) {
	hostByte, err := GetHostByteFromString(addr.String())
	if err != nil {
		return nil, err
	}
	var toClientBytes []byte
	toClientBytes = append(toClientBytes, UdpForwardVersion...)
	toClientBytes = append(toClientBytes, 0x00)
	toClientBytes = append(toClientBytes, hostByte...)
	toClientBytes = append(toClientBytes, data...)
	return toClientBytes, nil
}

// GetHostByteFromString is defined to generate [type, addr, port¬]
func GetHostByteFromString(hostStr string) ([]byte, error) {
	var buf []byte
	addr, port, err := net.SplitHostPort(hostStr)
	if err != nil {
		return nil, nil
	}
	if ip := net.ParseIP(addr); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buf = append(buf, AddressTypeIpv4)
			buf = append(buf, ip4...)
		} else {
			// ipv6
			buf = append(buf, AddressTypeIpv6)
			buf = append(buf, ip...)
		}
	} else {
		buf = append(buf, AddressTypeDomain)
		buf = append(buf, byte(len(addr)))
		buf = append(buf, []byte(addr)...)
	}

	portUint, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		return nil, err
	}
	portBytes := binary.BigEndian.AppendUint16([]byte{}, uint16(portUint))
	buf = append(buf, portBytes...)

	return buf, nil
}
