package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/url"

	"encoding/base64"
	"time"

	socks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"
	"github.com/launchdarkly/go-ntlm-proxy-auth"
)

var encBase64 = base64.StdEncoding.EncodeToString
var decBase64 = base64.StdEncoding.DecodeString
var username string = ""
var domain string = ""
var password string = ""
var connectproxystring string
var useragent string
var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response

func getProxyConnection(proxyaddr string, connectaddr string) net.Conn {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	proxyurl, err := url.Parse(proxyaddr)
	if err != nil {
		log.Println("url parse error:", err)
		return nil
	}

	ntlmDialContext := ntlm.NewNTLMProxyDialContext(dialer, *proxyurl, username, password, domain, nil)
	if ntlmDialContext == nil {
		log.Println("ntlmDialErr")
		return nil
	}
	ctx := context.Background()
	conn, err := ntlmDialContext(ctx, "tcp", connectaddr)
	if err != nil {
		log.Println("ntlm dialContext connection error:", err)
		return nil
	}

	return conn
}

func connectToServer(address string, proxy string) error {
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return err
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	var conn net.Conn
	var connp net.Conn
	var newconn net.Conn

	if proxy == "" {
		log.Println("Connecting to far end")
		conn, err = tls.Dial("tcp", address, conf)
		if err != nil {
			return err
		}
	} else {
		log.Println("Connecting to proxy ...")
		connp = getProxyConnection(proxy, address)
		if connp != nil {
			log.Println("Proxy connection successful. Connecting to far end...")
			conntls := tls.Client(connp, conf)
			err := conntls.Handshake()
			if err != nil {
				log.Printf("Error connecting: %v", err)
				return err
			}
			newconn = net.Conn(conntls)
		} else {
			log.Println("Proxy connection NOT successful. Exiting")
			return nil
		}
	}

	log.Println("Starting client")
	if proxy == "" {
		conn.Write([]byte(agentpassword))
		session, err = yamux.Server(conn, nil)
	} else {
		newconn.Write([]byte(agentpassword))
		time.Sleep(time.Second * 1)
		session, err = yamux.Server(newconn, nil)
	}
	if err != nil {
		return err
	}

	for {
		stream, err := session.Accept()
		log.Println("Acceping stream")
		if err != nil {
			return err
		}
		log.Println("Passing off to SOCKS")
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
