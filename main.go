package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var agentpassword string

func main() {

	// Server parameters
	serverFlags := flag.NewFlagSet("server", 0)
	listen := serverFlags.String("listen", "0.0.0.0:8080", "Listen address for client connections")
	certificate := serverFlags.String("cert", "server", "Server certificate file prefix")
	socks := serverFlags.String("socks", "127.0.0.1:1080", "Listen address for the SOCKS5 proxy")

	// Client parameters
	clientFlags := flag.NewFlagSet("client", 0)
	connect := clientFlags.String("connect", "", "address:port of the server to connect to")
	proxy := clientFlags.String("proxy", "", "URI of the proxy to use to connect to the server [optional]")
	proxyauthstring := clientFlags.String("proxyauth", "", "Proxy authentication in the format [Domain/]Username:Password [optional]")
	optproxytimeout := clientFlags.Int("proxytimeout", 1, "Proxy response timeout in seconds [optional]")
	optuseragent := clientFlags.String("useragent", "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko", "User-Agent [optional]")
	recn := clientFlags.Int("recn", 3, "Reconnection limit, 0 for infinite [optional]")
	rect := clientFlags.Int("rect", 30, "Reconnection delay [optional]")

	// Shared parameters
	version := flag.Bool("version", false, "Version information")
	optpassword := flag.String("pass", "", "Shared server/client password [optional]")

	serverFlags.Usage = func() {
		fmt.Println("SERVER MODE:")
		fmt.Printf("%s server [-listen <listenAddr>] [-socks <socksAddr>] [options]\n", os.Args[0])
		fmt.Println("Options:")
		serverFlags.PrintDefaults()
		fmt.Println()
	}

	clientFlags.Usage = func() {
		fmt.Println("CLIENT MODE:")
		fmt.Printf("%s client -connect <connectAddr> [-proxy <proxyURI>] [options]\n", os.Args[0])
		fmt.Println("Options:")
		clientFlags.PrintDefaults()
		fmt.Println()
	}

	flag.Usage = func() {
		fmt.Println("rsockstun - reverse socks5 server/client")
		fmt.Println()
		serverFlags.Usage()
		clientFlags.Usage()
		fmt.Println("General Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Note: you can generate a new server certificate with the following command:")
		fmt.Println("openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes")
	}

	flag.Parse()

	if *version {
		fmt.Println("rsockstun - reverse socks5 server/client")
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Defaults

	proxytimeout = time.Duration(1000 * int(*optproxytimeout))

	if *optpassword != "" {
		agentpassword = *optpassword
	} else {
		agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
	}

	useragent = *optuseragent

	switch strings.ToLower(flag.Arg(0)) {

	case "server":
		// Server Mode
		serverFlags.Parse(os.Args[2:])
		go clientListener(*listen, *certificate)
		log.Fatal(socksListener(*socks))

	case "client":
		// Client Mode
		clientFlags.Parse(os.Args[2:])
		if *connect == "" {
			fmt.Println("Please specify a connect address!")
			os.Exit(1)
		}

		if *proxyauthstring != "" {
			userpass := *proxyauthstring
			if strings.Contains(strings.Split(userpass, ":")[0], "/") {
				domain = strings.Split(userpass, "/")[0]
				userpass = strings.Split(userpass, "/")[1]
			}
			username = strings.Split(userpass, ":")[0]
			password = strings.Split(userpass, ":")[1]
		}

		if *recn > 0 {
			for i := 1; i <= *recn; i++ {
				log.Printf("Connecting to the far end. Try %d of %d", i, *recn)
				error1 := connectToServer(*connect, *proxy)
				log.Print(error1)
				log.Printf("Sleeping for %d sec...", *rect)
				tsleep := time.Second * time.Duration(*rect)
				time.Sleep(tsleep)
			}
		} else {
			for {
				log.Printf("Reconnecting to the far end... ")
				error1 := connectToServer(*connect, *proxy)
				log.Print(error1)
				log.Printf("Sleeping for %d sec...", *rect)
				tsleep := time.Second * time.Duration(*rect)
				time.Sleep(tsleep)
			}
		}
		log.Fatal("Ending...")
	default:
		fmt.Fprintf(os.Stderr, "You must specify a mode between \"server\" and \"client\".\n")
		os.Exit(1)
	}
}
