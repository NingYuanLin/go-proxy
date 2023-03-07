package socks5

import (
	"bytes"
	"net"
	"reflect"
	"testing"
)

func TestNewClientRequestMessage(t *testing.T) {
	tests := []struct {
		name string

		Version     byte
		Cmd         byte
		Reserved    byte
		AddressType byte
		Address     []byte
		Port        []byte

		Err     error
		Message ClientRequestMessage
	}{
		{
			"valid connect-ipv4 test",
			Socks5Version,
			CmdConnect,
			ReversedField,
			AddressTypeIpv4,
			[]byte{1, 1, 1, 1},
			[]byte{0x00, 0x50},
			nil,
			ClientRequestMessage{
				CmdConnect,
				"1.1.1.1",
				80,
				AddressTypeIpv4,
			},
		},
		{
			"invalid version test",
			0x00,
			CmdConnect,
			ReversedField,
			AddressTypeIpv4,
			[]byte{1, 1, 1, 1},
			[]byte{0x00, 0x50},
			ErrVersionNotSupport,
			ClientRequestMessage{
				CmdConnect,
				"1.1.1.1",
				80,
				AddressTypeIpv4,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.Buffer{}
		buf.Write([]byte{test.Version, test.Cmd, test.Reserved, test.AddressType})
		buf.Write(test.Address)
		buf.Write(test.Port)

		message, err := NewClientRequestMessage(&buf)
		if err != test.Err {
			t.Fatalf("should get error %s, but got %s", test.Err, err)
		}
		if err != nil {
			continue
		}

		if *message != test.Message {
			t.Fatalf("should get clientRequestMessage = %v, but got %v", message, test.Message)
		}

	}
}

func TestWriteRequestSuccessReply(t *testing.T) {
	tests := []struct {
		Conn *bytes.Buffer
		Ip   net.IP
		Port uint16

		Message []byte
		Err     error
	}{
		{
			&bytes.Buffer{},
			net.IP([]byte{1, 1, 1, 1}),
			uint16(80),

			[]byte{Socks5Version, ReplySuccess, ReversedField, AddressTypeIpv4, 1, 1, 1, 1, 0x00, 0x50},
			nil,
		},
		{
			&bytes.Buffer{},
			net.ParseIP("2002:1::1"),
			uint16(80),

			[]byte{Socks5Version, ReplySuccess, ReversedField, AddressTypeIpv4, 1, 1, 1, 1, 0x00, 0x50},
			nil,
		},
	}

	for _, test := range tests {
		err := WriteRequestSuccessReply(test.Conn, test.Ip, test.Port)
		if err != test.Err {
			t.Fatalf("error should be %s, but got %s", test.Err, err)
		}
		if err != nil {
			return
		}

		message := test.Conn.Bytes()
		if !reflect.DeepEqual(message, test.Message) {
			t.Fatalf("message should be %v, but got %v", test.Message, message)
		}
	}
}

func TestWriteRequestFailureReply(t *testing.T) {
	tests := []struct {
		Conn      *bytes.Buffer
		ReplyType ReplyType

		Message []byte
		Err     error
	}{
		{
			&bytes.Buffer{},
			ReplyConnectionNotAllowed,
			[]byte{Socks5Version, ReplyConnectionNotAllowed, ReversedField, AddressTypeIpv4, 0, 0, 0, 0, 0, 0},
			nil,
		},
	}

	for _, test := range tests {
		err := WriteRequestFailureReply(test.Conn, test.ReplyType)
		if err != test.Err {
			t.Fatalf("error should be %s, but got %s", test.Err, err)
		}
		if err != nil {
			return
		}

		message := test.Conn.Bytes()
		if !reflect.DeepEqual(message, test.Message) {
			t.Fatalf("message should be %v, but got %v", test.Message, message)
		}
	}
}
