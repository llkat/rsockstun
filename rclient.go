package main

import (
	"log"
	"net"
	"crypto/tls"

	socks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"
	"encoding/base64"
	"time"
	"net/http"
	"bufio"
	"strings"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
	"io/ioutil"
)


var encBase64 = base64.StdEncoding.EncodeToString
var decBase64 = base64.StdEncoding.DecodeString
var username string
var domain string
var password string
var connectproxystring string
var useragent string
var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response



func connectviaproxy(proxyaddr string, connectaddr string) net.Conn {

	if (username != "") && (password != "") && (domain != "") {
		connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
			"\r\nUser-Agent: "+useragent+
			"\r\nProxy-Authorization: NTLM TlRMTVNTUAABAAAABoIIAAAAAAAAAAAAAAAAAAAAAAA=" +
			"\r\nProxy-Connection: Keep-Alive" +
			"\r\n\r\n"

	}else	{
		connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
			"\r\nUser-Agent: "+useragent+
			"\r\nProxy-Connection: Keep-Alive" +
			"\r\n\r\n"
	}

	//log.Print(connectproxystring)

	conn, err := net.Dial("tcp", proxyaddr)
	if err != nil {
		// handle error
		log.Printf("Error connect: %v",err)
	}
	conn.Write([]byte(connectproxystring))

	time.Sleep(proxytimeout) //Because socket does not close - we need to sleep for full response from proxy

	resp,err := http.ReadResponse(bufio.NewReader(conn),&http.Request{Method: "CONNECT"})
	status := resp.Status

	//log.Print(status)
	//log.Print(resp)

	if (resp.StatusCode == 200)  || (strings.Contains(status,"HTTP/1.1 200 ")) ||
		(strings.Contains(status,"HTTP/1.0 200 ")){
		log.Print("Connected via proxy. No auth required")
		return conn
	}

	if (strings.Contains(status,"407 Proxy Authentication Required")){
		log.Print("Got Proxy auth:")
		ntlmchall := resp.Header.Get("Proxy-Authenticate")
		if ntlmchall != "" {
			log.Print("Got NTLM challenge:")
			//log.Print(ntlmchall)
			var session ntlm.ClientSession
			session, _ = ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionlessMode)
			session.SetUserInfo(username,password,domain)
			//negotiate, _ := session.GenerateNegotiateMessage()
			//log.Print(negotiate)

			ntlmchall = ntlmchall[5:]
			ntlmchallb,_ := decBase64(ntlmchall)


			challenge, _ := ntlm.ParseChallengeMessage(ntlmchallb)
			session.ProcessChallengeMessage(challenge)
			authenticate, _ := session.GenerateAuthenticateMessage()
			ntlmauth:= encBase64(authenticate.Bytes())

			//log.Print(authenticate)
			connectproxystring = "CONNECT "+connectaddr+" HTTP/1.1"+"\r\nHost: "+connectaddr+
				"\r\nUser-Agent: Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko"+
				"\r\nProxy-Authorization: NTLM "+ntlmauth+
				"\r\nProxy-Connection: Keep-Alive"+
				"\r\n\r\n"


			//Empty read buffer
			/*
			var statusb []byte
			//conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
			bufReader := bufio.NewReader(conn)
			n, err := bufReader.Read(statusb)
			//statusb,_ := ioutil.ReadAll(bufReader)
			if err != nil {
				if err == io.EOF {
					log.Printf("Readed %v vites",n)
				}
			}
			status = string(statusb)
			*/

			conn.Write([]byte(connectproxystring))

			//read response
			bufReader := bufio.NewReader(conn)
			conn.SetReadDeadline(time.Now().Add(proxytimeout))
			statusb,_ := ioutil.ReadAll(bufReader)

			status = string(statusb)

			//disable socket read timeouts
			conn.SetReadDeadline(time.Now().Add(100 * time.Hour))

			if (strings.Contains(status,"HTTP/1.1 200 ")){
				log.Print("Connected via proxy")
				return conn
			} else{
				log.Printf("Not Connected via proxy. Status:%v",status)
				return nil
			}
		}

	}else {
		log.Print("Not connected via proxy")
		conn.Close()
		return nil
	}

	return conn
}

func connectForSocks(address string, proxy string) error {
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
	//var conntls tls.Conn
	//var conn tls.Conn
	if proxy == "" {
		log.Println("Connecting to far end")
		//conn, err = net.Dial("tcp", address)
		conn, err = tls.Dial("tcp", address, conf)
		if err != nil {
			return err
		}
	}else {
		log.Println("Connecting to proxy ...")
		connp = connectviaproxy(proxy,address)
		if connp != nil{
			log.Println("Proxy successfull. Connecting to far end")
			conntls := tls.Client(connp,conf)
			err := conntls.Handshake()
			if err != nil {
				log.Printf("Error connect: %v",err)
				return err
			}
			newconn = net.Conn(conntls)
		}else{
			log.Println("Proxy NOT successfull. Exiting")
			return nil
		}
	}

	log.Println("Starting client")
	if proxy == "" {
		conn.Write([]byte(agentpassword))
		//time.Sleep(time.Second * 1)
		session, err = yamux.Server(conn, nil)
	}else {

		//log.Print(conntls)
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
		log.Println("Passing off to socks5")
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
