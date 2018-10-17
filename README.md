rsockstun
======

Reverse socks5 tunneler with SSL and proxy support
Based on https://github.com/brimstone/rsocks

Usage:
------
```

Usage:
0) Generate self-signed certificate with openssl: openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes
1) Start on VPS: rsockstun -listen :8443 -socks 127.0.0.1:1080 -cert cert  
2) Start on client: rsockstun -connect clientIP:8443
3) Connect to 127.0.0.1:1080 on the VPS with any socks5 client.
4) Enjoy. :]

Add params:
 -proxy 1.2.3.4:3128 - connect via proxy
 -proxyauth Domain/username:password  - proxy credendials
 -proxytimeout 2000 - server and clients will wait for 2000 msec for proxy connections... (Sometime it should be up to 4000...)
 -useragent - User-Agent string
 

in linux VPS 
install Golang: apt install golang
export GOPATH=~/go
go get github.com/hashicorp/yamux
go get github.com/armon/go-socks5
go get github.com/ThomsonReutersEikon/go-ntlm/ntlm
go build

in windows client:
download and install golang

go get github.com/hashicorp/yamux
go get github.com/armon/go-socks5
go get github.com/ThomsonReutersEikon/go-ntlm/ntlm
If you want to use proxy NTLM auth - patch go-ntlm\ntlm\payload.go packet:
	bytes := utf16FromString(value) -> bytes := []byte(value)
	p.Type = UnicodeStringPayload   -> p.Type = OemStringPayload
go build

By default client connects to server and send "magic bytes" to authorize on server. If server does not receive magic bytes from client (for example if spider or client browser connects to server ) then it send HTTP 301 redirect code to www.microsoft.com

```