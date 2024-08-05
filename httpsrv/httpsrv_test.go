package httpsrv

import (
	"bufio"
	"context"
	"github.com/pires/go-proxyproto"
	"log"
	"net"
	"net/http"
	"testing"
)

func chkErr(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}
}

func TestClientProxyproto(t *testing.T) {
	// Dial some proxy listener e.g. https://github.com/mailgun/proxyproto
	target, err := net.ResolveTCPAddr("tcp", "127.0.0.1:2319")
	chkErr(err)

	conn, err := net.DialTCP("tcp", nil, target)
	chkErr(err)

	defer conn.Close()

	// Create a proxyprotocol header or use HeaderProxyFromAddrs() if you
	// have two conn's
	header := &proxyproto.Header{
		Version:           2,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.TCPv4,
		SourceAddr: &net.TCPAddr{
			IP:   net.ParseIP("10.1.1.1"),
			Port: 1000,
		},
		DestinationAddr: &net.TCPAddr{
			IP:   net.ParseIP("20.2.2.2"),
			Port: 2000,
		},
	}
	// After the connection was created write the proxy headers first
	_, err = header.WriteTo(conn)
	chkErr(err)

	// Create an HTTP request
	request := "GET /whoami HTTP/1.1\r\n" +
		"Host: 127.0.0.1\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	// Write the HTTP request to the TCP connection
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(request)
	if err != nil {
		chkErr(err)

	}

	// Flush the buffer to ensure the request is sent
	err = writer.Flush()
	if err != nil {
		chkErr(err)
	}
}

func TestServiceProxyproto(t *testing.T) {
	s, err := NewService(":2319", WithHandler(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		log.Printf("[Handler] remote ip %q", r.RemoteAddr)
	})), WithEnableProxyProtocol(true))
	if err != nil {
		t.Fatal(err)
	}

	s.Start(context.Background())
}
