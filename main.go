package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/yamux"
	"time"

	"strconv"
	"strings"
)

var session *yamux.Session
var agentpassword string

func main() {

	listen := flag.String("listen", "", "listen port for receiver address:port")
	certificate := flag.String("cert", "", "certificate file")
	socks := flag.String("socks", "127.0.0.1:1080", "socks address:port")
	connect := flag.String("connect", "", "connect address:port")
	proxy := flag.String("proxy", "", "proxy address:port")
	optproxytimeout := flag.String("proxytimeout", "", "proxy response timeout (ms)")
	proxyauthstring := flag.String("proxyauth", "", "proxy auth Domain/user:Password ")
	optuseragent := flag.String("useragent", "", "User-Agent")
	optpassword := flag.String("agentpassword", "", "connect password")
	version := flag.Bool("version", false, "version information")
	flag.Usage = func() {
		fmt.Println("rsockstun - reverse socks5 server/client")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("1) Start rsockstun -listen :8080 -socks 127.0.0.1:1080 on the client.")
		fmt.Println("2) Start rsockstun -connect client:8080 on the server.")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the client with any socks5 client.")
		fmt.Println("4) Start rsockstun -connect client:8080 -proxy 1.2.3.4:3124 -proxyauth Domain/user:pass")
		fmt.Println("X) Enjoy. :]")
	}

	flag.Parse()

	if *version {
		fmt.Println("rsockstun - reverse socks5 server/client")
		os.Exit(0)
	}

	if *listen != "" {
		log.Println("Starting to listen for clients")
		if *optproxytimeout != "" {
			opttimeout,_ := strconv.Atoi(*optproxytimeout)
			proxytout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytout = time.Millisecond * 1000
		}

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		go listenForSocks(*listen, *certificate)
		log.Fatal(listenForClients(*socks))
	}

	if *connect != "" {
		log.Println("Connecting to the far end")

		if *optproxytimeout != "" {
			opttimeout,_ := strconv.Atoi(*optproxytimeout)
			proxytimeout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytimeout = time.Millisecond * 1000
		}

		if *proxyauthstring != "" {
			domain = strings.Split(*proxyauthstring, "/")[0]
			username = strings.Split(strings.Split(*proxyauthstring, "/")[1],":")[0]
			password = strings.Split(strings.Split(*proxyauthstring, "/")[1],":")[1]
		} else {
			domain = ""
			username = ""
			password = ""
		}

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		if *optuseragent != "" {
			useragent = *optuseragent
		} else {
			useragent = "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko"
		}
		log.Fatal(connectForSocks(*connect,*proxy))
	}

	fmt.Fprintf(os.Stderr, "You must specify a listen port or a connect address")
	os.Exit(1)
}
