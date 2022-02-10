package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
)

var session *yamux.Session

var proxytout = time.Millisecond * 1000 //timeout for wait magicbytes
// Catches yamux connecting to us
func clientListener(address string, certificate string) {
	log.Printf("Listening for clients on %v\n", address)

	cer, err := tls.LoadX509KeyPair(certificate+".crt", certificate+".key")
	if err != nil {
		log.Println(err)
		fmt.Fprintf(os.Stderr, "Please check the program's usage on how to generate a new SSL certificate.\n")
		os.Exit(1)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", address, config)
	if err != nil {
		return
	}
	for {
		conn, err := ln.Accept()
		conn.RemoteAddr()
		log.Printf("Got a SSL connection from %v: ", conn.RemoteAddr())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Errors accepting connection!")
		}

		reader := bufio.NewReader(conn)

		//read only 64 bytes with timeout=1-3 sec. So we haven't delay with browsers
		conn.SetReadDeadline(time.Now().Add(proxytout))
		statusb := make([]byte, 64)
		_, _ = io.ReadFull(reader, statusb)

		if string(statusb)[:len(agentpassword)] != agentpassword {
			//do HTTP checks
			log.Printf("Received request: %v", string(statusb[:64]))
			status := string(statusb)
			if strings.Contains(status, " HTTP/1.1") {
				httpresonse := "HTTP/1.1 301 Moved Permanently" +
					"\r\nContent-Type: text/html; charset=UTF-8" +
					"\r\nLocation: https://www.microsoft.com/" +
					"\r\nServer: Apache" +
					"\r\nContent-Length: 0" +
					"\r\nConnection: close" +
					"\r\n\r\n"

				conn.Write([]byte(httpresonse))
				conn.Close()
			} else {
				conn.Close()
			}

		} else {
			//magic bytes received.
			//disable socket read timeouts
			log.Println("Client Connected.")
			conn.SetReadDeadline(time.Now().Add(100 * time.Hour))

			//Add connection to yamux
			session, err = yamux.Client(conn, nil)
		}
	}
}

// Catches clients and connects to yamux
func socksListener(address string) error {
	log.Printf("Listening for SOCKS connections on %v\n", address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		if session == nil {
			conn.Close()
			continue
		}
		log.Println("Got a connection, opening a stream...")

		stream, err := session.Open()
		if err != nil {
			return err
		}

		// connect both of conn and stream
		go func() {
			log.Println("Starting to copy conn to stream")
			io.Copy(conn, stream)
			conn.Close()
		}()
		go func() {
			log.Println("Starting to copy stream to conn")
			io.Copy(stream, conn)
			stream.Close()
			log.Println("Done copying stream to conn")
		}()
	}
}
