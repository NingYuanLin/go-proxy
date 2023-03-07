package main

import (
	"github.com/NingYuanLin/go-proxy/socks5"
	"github.com/spf13/cobra"
	"log"
	"net"
	"time"
)

var startCmd = &cobra.Command{
	Use: "start",
	Short: "To start socks5 server",
	Run: func(cmd *cobra.Command, args []string) {
		configFromFile, err := parseConfigFromFile()

		if err != nil {
			log.Panicln(err)
		}
		ip := configFromFile.Ip
		port := configFromFile.port
		username := configFromFile.username
		password := configFromFile.password
		udpRelayServerIp := configFromFile.udp_relay_server_ip
		udpPort := configFromFile.udp_port
		timeout := configFromFile.timeout
		udpConnLifetime := configFromFile.udp_conn_lifetime

		var passwordChecker socks5.PasswordCheckerFunc
		if username != "" && password != "" {
			passwordChecker = func(uname, pwd string) bool {
				if username == uname && password == pwd {
					return true
				}
				return false
			}
		}

		var authMethod socks5.AuthMethod = socks5.MethodNoAuth
		if passwordChecker != nil {
			authMethod = socks5.MethodPassword
		}

		socks5Server := socks5.Socks5Server{
			Ip:   ip,
			Port: port,
			Config: socks5.Config{
				AuthMethod:       authMethod,
				Timeout:          time.Second * time.Duration(timeout),
				PasswordChecker:  passwordChecker,
				UdpRelayServerIp: net.ParseIP(udpRelayServerIp),
				UdpPort:          udpPort,
				UdpConnLifetime:  time.Second * time.Duration(udpConnLifetime),
			},
		}

		log.Println("start server")
		err = socks5Server.Run()
		if err != nil {
			log.Println(err)
		}
	},
}
