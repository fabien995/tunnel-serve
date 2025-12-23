package client

import (
	"github.com/hashicorp/yamux"
	"net"
	"io"
	"strings"
	"log"
)

func Run(outboundTunnelTarget string, routeTarget string) {
	client(outboundTunnelTarget, routeTarget)
}


/*
	Main client function.
	- outboundTunnelTarget: address of the reverse tunnel server
	- routeTarget: address of the local service which should be made available
*/
func client(outboundTunnelTarget string, routeTarget string) {
	var tag string = "[test-mux-client-client] -"

	conn, err := net.Dial("tcp", outboundTunnelTarget)
	if err != nil {
		panic(err)
	}

	// Yamux server functionality on client-side
	// for reverse tunneling.
	session, err := yamux.Server(conn, nil)
	if err != nil {
		panic(err)
	}

	log.Printf("%s client connected to yamux session.\n", tag)
	log.Printf("%s client target of tunnel: %s.\n", tag, outboundTunnelTarget)

	
	controlInfoConn, err := session.Accept()
	if err != nil {
		panic(err)
	}
	// Send secret so that server knows
	// this is a valid client.
	secret := "d12a1f29-065d-4d65-addf-fefa51ff019b"
	sendSecret := []byte(secret)
	controlInfoConn.Write(sendSecret)
	log.Printf("%s Sent auth secret.\n", tag)

	// Get control info
	// (address of gateway).
	addrRecv := make([]byte, 80)
	controlInfoConn.Read(addrRecv)
	addrRecvStr := string(addrRecv)
	log.Printf("%s control msg: %s\n", tag, addrRecvStr)
	addrRecvStr = decode(addrRecvStr)
	log.Printf("%s Gateway address: %s.\n", tag, addrRecvStr)
	controlInfoConn.Close()

	for {
		stream, err := session.Accept()
		if err != nil {
			panic(err)
		}

		go copyConnection(stream, routeTarget)
	}
}

func copyConnection(stream net.Conn, routeTarget string) {
	var tag string = "[test-mux-copyConnection] -"

	log.Printf("%s client session accepted.\n", tag)

	// Destination for tunnel.
	conn, err := net.Dial("tcp", routeTarget)
	if err != nil {
		panic(err)
	}


	defer conn.Close()
	defer stream.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(conn, stream)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(stream, conn)
		done <- struct{}{}
	}()

	<-done
}

func decode(paddedStr string) string {
	splitStrs := strings.Split(paddedStr, ",")
	return splitStrs[1]
}
