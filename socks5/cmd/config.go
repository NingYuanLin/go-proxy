package main

import (
	"bufio"
	"fmt"
	"github.com/NingYuanLin/go-proxy/socks5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
)

var configCmd = &cobra.Command{
	Use: "config",
	Short: "The operations about config file",
	Run: func(cmd *cobra.Command, args []string) {
		if create, _ := cmd.Flags().GetBool("create"); create == true {
			// create config file
			err := createConfigFile()
			if err != nil {
				log.Println(err)
			}
			return
		}
	},
}

func init() {
	configCmd.Flags().Bool("create", false, "create config file")
	configCmd.MarkFlagRequired("create")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	//fmt.Println("initConfig")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		//print(home)
		cobra.CheckErr(err)

		// Search config in home directory with name "go-proxy.yaml".
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("go-proxy")
	}

	//viper.AutomaticEnv() // read in environment variables that match

}

func readConfigToViper() (configFileUsed string, err error) {
	// If a config file is found, read it in.
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	configFileUsed = viper.ConfigFileUsed()
	return
}

func createConfigFile() error {
	// check if the config file exist and read to viper
	configFile, err := readConfigToViper()
	exist := false
	if err == nil {
		exist = true
	}

	if exist {
		//	log.Println("Using config file:", configFile)
		log.Printf("Warning: Config file is existed in %s, and the following operation will rewrite the config file.\n", configFile)
	}

	fmt.Print("Please input listen ip(default: 0.0.0.0): ")
	reader := bufio.NewReader(os.Stdin)
	ip, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "0.0.0.0"
	}

	fmt.Print("Please input listen port(default: 1080): ")
	port, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	port = strings.TrimSpace(port)
	portInt := 1080
	if port != "" {
		portInt64, err := strconv.ParseInt(port, 10, 16)
		if err != nil {
			return err
		}
		portInt = int(portInt64)
	}

	usePasswordAuth := false
	for {
		fmt.Print("Do you want to set username and password? (Y/n) (default: n): ")
		state, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		state = strings.TrimSpace(state)
		state = strings.ToLower(state)
		if state == "" || state == "n" {
			break
		}
		if state == "y" {
			usePasswordAuth = true
			break
		}
	}

	var username string
	var password string
	if usePasswordAuth == true {
		for username == "" {
			fmt.Print("Please input username: ")
			username, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			username = strings.TrimSpace(username)
		}

		for password == "" {
			fmt.Print("Please input password: ")
			password, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			password = strings.TrimSpace(password)
		}
	}

	openUdpRelay := false
	for {
		fmt.Print("Do you want to open udp relay? (Y/n) (default:n): ")
		state, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		state = strings.TrimSpace(state)
		state = strings.ToLower(state)
		if state == "" || state == "n" {
			break
		}
		if state == "y" {
			openUdpRelay = true
			break
		}
	}

	var udpRelayServerIp string
	var udpPort socks5.UdpRelayPort = socks5.UdpRelayClose
	if openUdpRelay == true {
		fmt.Println("By default, the udp relay server ip will be detected automatically.")
		fmt.Println("If your relay server is under NAT, you may need to set udp relay server ip as your server's public ip manually.")
		fmt.Print("Please input your server ip (default: auto): ")
		udpIp, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		udpRelayServerIp = strings.TrimSpace(udpIp)

		for {
			fmt.Println("You can set a fixed port as udp listen port, such as same as tcp port you set before. However, it is not safe. The attacker can jump username-password auth process and access it directly.")
			fmt.Println("We strongly suggest you to use random udp port. Don't forget to open your firewall to allow all udp access from any ports.")
			fmt.Printf("Please input your udp listen port (default: same as tcp port(%d). 0: random port. 1~65535: fixed port.)", portInt)
			udpPortStr, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			udpPortStr = strings.TrimSpace(udpPortStr)
			if udpPortStr == "" {
				udpPort = portInt
			} else if udpPortStr == "0" {
				udpPort = socks5.UdpRelayRandomPort
			} else {
				parseInt, err := strconv.ParseInt(udpPortStr, 10, 16)
				if err != nil {
					log.Println(err)
					continue
				}
				udpPort = int(parseInt)
			}
			break
		}

	}

	advancedSetting := false
	for {
		fmt.Print("Do you want to perform advanced setting? (Y/n) (default:n): ")
		state, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		state = strings.TrimSpace(state)
		state = strings.ToLower(state)
		if state == "" || state == "n" {
			break
		}
		if state == "y" {
			advancedSetting = true
			break
		}
	}

	// unit: seconds
	var timeout int64 = 3
	var udpConnLifetime int64 = 60
	if advancedSetting == true {
		fmt.Print("Please input the timeout of tcp dial (unit: seconds) (default: 3): ")
		timeoutStr, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		timeoutStr = strings.TrimSpace(timeoutStr)
		if timeoutStr != "" {
			parseInt, err := strconv.ParseInt(timeoutStr, 10, 64)
			if err != nil {
				return err
			}
			timeout = parseInt
		}

		if openUdpRelay == true {
			fmt.Print("please input the lifetime of udp exchange socket (unit: seconds) (default: 60): ")
			udpLifeTimeStr, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			udpLifeTimeStr = strings.TrimSpace(udpLifeTimeStr)
			if udpLifeTimeStr != "" {
				parseInt, err := strconv.ParseInt(udpLifeTimeStr, 10, 64)
				if err != nil {
					return err
				}
				udpConnLifetime = parseInt
			}
		}
	}

	viper.Set("ip", ip)
	viper.Set("port", portInt)
	viper.Set("username", username)
	viper.Set("password", password)
	viper.Set("udp_relay_server_ip", udpRelayServerIp)
	viper.Set("udp_port", udpPort)
	viper.Set("timeout", timeout)
	viper.Set("udp_conn_lifetime", udpConnLifetime)

	if exist == true {
		err = viper.WriteConfig()
	} else {
		err = viper.SafeWriteConfig()
	}
	if err != nil {
		return err
	}

	return nil
}

type ConfigFileStruct struct {
	Ip                  string
	port                int
	username            string
	password            string
	udp_relay_server_ip string
	udp_port            int
	timeout             int64
	udp_conn_lifetime   int64
}

func parseConfigFromFile() (*ConfigFileStruct, error) {
	configFile, err := readConfigToViper()
	if err != nil {
		return nil, err
	}
	log.Println("Using config file:", configFile)

	configFileStruct := &ConfigFileStruct{}
	configFileStruct.Ip = viper.GetString("ip")
	configFileStruct.port = viper.GetInt("port")
	configFileStruct.username = viper.GetString("username")
	configFileStruct.password = viper.GetString("password")
	configFileStruct.udp_relay_server_ip = viper.GetString("udp_relay_server_ip")
	configFileStruct.udp_port = viper.GetInt("udp_port")
	configFileStruct.timeout = viper.GetInt64("timeout")
	configFileStruct.udp_conn_lifetime = viper.GetInt64("udp_conn_lifetime")

	err = viper.Unmarshal(configFileStruct)
	if err != nil {
		return nil, err
	}

	return configFileStruct, nil
}
