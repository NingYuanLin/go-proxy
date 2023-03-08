# Go-proxy
Standard and lightweight proxy implementation with golang.
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
