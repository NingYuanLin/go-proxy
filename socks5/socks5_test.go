package socks5

import (
	"testing"
)

func TestSocks5NoAuthServer(t *testing.T) {
	ip := "0.0.0.0"
	port := 333
	config := Config{
		AuthMethod: MethodNoAuth,
		UdpPort:    port,
	}
	server := Socks5Server{
		Ip:     ip,
		Port:   port,
		Config: config,
	}

	err := server.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSocks5PasswordAuthServer(t *testing.T) {
	ip := "0.0.0.0"
	port := 333
	config := Config{
		AuthMethod: MethodPassword,
		PasswordChecker: func(username, password string) bool {
			if username != "123" || password != "456" {
				return false
			}
			return true
		},
		UdpPort: port,
	}
	server := Socks5Server{
		Ip:     ip,
		Port:   port,
		Config: config,
	}

	err := server.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSocks5RandomUdpPortServer(t *testing.T) {
	ip := "0.0.0.0"
	port := 333
	config := Config{
		AuthMethod: MethodNoAuth,
		UdpPort:    UdpRelayRandomPort,
	}
	server := Socks5Server{
		Ip:     ip,
		Port:   port,
		Config: config,
	}

	err := server.Run()
	if err != nil {
		t.Fatal(err)
	}
}
