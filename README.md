# Go-proxy
Standard and lightweight proxy implementations with golang.
* Socks5
## Socks5-server
SOCKS Protocol Version 5 Library.
* Both TCP/UDP supported
* No-auth and password-auth methods supported
* Fixed port and random port supported for udp relay.

### Use as Cli tool
#### 1. install
```
go install github.com/NingYuanLin/go-proxy/socks5/socks5-cmd@latest
```
#### 2. Create config file
```
socks5-cmd config --create 
```
```
base ❯ socks5-cmd config --create
2023/04/17 21:20:37 Warning: Config file is existed in /Users/ning/go-proxy.yaml, and the following operation will rewrite the config file.
Please input listen ip(default: 0.0.0.0): 0.0.0.0
Please input listen port(default: 1080): 1080
Do you want to set username and password? (Y/n) (default: n): y
Please input username: 123
Please input password: 456
Do you want to open udp relay? (Y/n) (default:n): y
By default, the udp relay server ip will be detected automatically.
If your relay server is under NAT, you may need to set udp relay server ip as your server's public ip manually.
Please input your server ip (default: auto):
You can set a fixed port as udp listen port, such as same as tcp port you set before. However, it is not safe. The attacker can jump username-password auth process and access it directly.
We strongly suggest you to use random udp port. Don't forget to open your firewall to allow all udp access from any ports.
Please input your udp listen port (default: same as tcp port(1080). 0: random port. 1~65535: fixed port.)0
Do you want to perform advanced setting? (Y/n) (default:n): y
Please input the timeout of tcp dial (unit: seconds) (default: 3): 3
please input the lifetime of udp exchange socket (unit: seconds) (default: 60): 60
```
#### 3. Start
```
socks5-cmd start
```
> If you want to run it in the background, you can use "nohup".

### Use as library
```
go get https://github.com/NingYuanLin/go-proxy.git@latest
```
### Thanks
* https://www.rfc-editor.org/rfc/rfc1928
* https://www.rfc-editor.org/rfc/rfc1929
* https://github.com/txthinking/socks5
* https://github.com/jqqjj/socks5
* https://github.com/0990/socks5
