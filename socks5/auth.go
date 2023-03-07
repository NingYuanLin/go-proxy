package socks5

import (
	"errors"
	"io"
)

type AuthMethod = byte

const (
	MethodNoAuth       AuthMethod = 0x00
	MethodGssApi       AuthMethod = 0x01
	MethodPassword     AuthMethod = 0x02
	MethodNoAcceptable AuthMethod = 0xff
)

const (
	PasswordAuthVersion = 0x01
	PasswordAuthSuccess = 0x00
	PasswordAuthFailure = 0x01
)

var ErrPasswordAuthFailure = errors.New("error authenticating password")

type ClientAuthMessage struct {
	NMethods byte
	Methods  []AuthMethod
}

func NewClientAuthMessage(conn io.Reader) (*ClientAuthMessage, error) {
	// Read version and nMethod
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}

	// validate version
	if buf[0] != Socks5Version {
		return nil, ErrVersionNotSupport
	}

	nMethods := buf[1]
	buf = make([]byte, nMethods)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}

	clientAuthMessage := ClientAuthMessage{
		NMethods: nMethods,
		Methods:  buf,
	}

	return &clientAuthMessage, nil
}

func WriteServerAuthMessage(conn io.Writer, method AuthMethod) error {
	buf := []byte{Socks5Version, method}
	_, err := conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

type ClientPasswordAuthMessage struct {
	Username string
	Password string
}

func NewClientPasswordAuthMessage(conn io.Reader) (*ClientPasswordAuthMessage, error) {
	// read version and usernameLen
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}
	version, usernameLen := buf[0], buf[1]

	if version != PasswordAuthVersion {
		return nil, ErrMethodVersionNotSupported
	}

	// read username and passwordLen
	buf = make([]byte, usernameLen+1)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}
	clientPasswordAuthMessage := ClientPasswordAuthMessage{}
	username := string(buf[:len(buf)-1])
	clientPasswordAuthMessage.Username = username

	passwordLen := buf[len(buf)-1]

	// read password
	if int(passwordLen) > len(buf) {
		buf = make([]byte, passwordLen)
	}
	_, err = io.ReadFull(conn, buf[:passwordLen])
	if err != nil {
		return nil, err
	}
	password := string(buf[:passwordLen])
	clientPasswordAuthMessage.Password = password

	return &clientPasswordAuthMessage, nil
}

func WriteServerPasswordMessage(conn io.Writer, status byte) error {
	_, err := conn.Write([]byte{PasswordAuthVersion, status})
	if err != nil {
		return err
	}
	return nil
}
