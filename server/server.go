package server

import (
	"github.com/hashicorp/yamux"
	"net"
	"io"
	"fmt"
	"errors"
	"log"
	"time"
	"strings"
)



/*
	Error handling for net.listen().accept().
	- err: Error,
	- lastErrorTime: last error time
	- consecutiveErrors: amount of errors within the error reset duration
	returns:
	- lastErrorTime,
	- consecutiveErrors
*/
func acceptErrorHandle(err error, lastErrorTime time.Time, consecutiveErrors int) (time.Time, int) {
	var tag string = "[server-acceptErrorHandle] -"
	
	var errorResetDuration = 10 * time.Second
	var maxConsecutiveErrors = 10

	if errors.Is(err, net.ErrClosed) {
		log.Printf("%s Got fatal err: %s\n", tag, err)
		panic(err)
	}

	log.Printf("%s Got error: %s\n", tag, err)

	// Error handling with max errors,
	// cooldown and reset of error counter
	// after a specific time

	// Reset counter if enough time has passed
	// since last error.
	if time.Since(lastErrorTime) > errorResetDuration {
		consecutiveErrors = 0
	}

	consecutiveErrors++
	lastErrorTime = time.Now()

	// If too many errors, shutdown
	if consecutiveErrors >= maxConsecutiveErrors {
		log.Printf("%s too many errors. Current err: %s\n", tag, err)
		panic(err)
	}

	return lastErrorTime, consecutiveErrors
}

/*
	Listen on control port
	for new clients.
	- bindAddress: IP address to listen on for control server and
		gateway server.
	- controlPort: port to bind to for control server.
*/
func muxServer(bindAddress string, controlPort int, domainName string) {
	var tag string = "[server-mux-server] -"

	listenAddress := fmt.Sprintf("%s:%d", bindAddress, controlPort)

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		panic(err)
	}

	log.Printf("%s muxServer listening on %s\n", tag, listenAddress)

	// Vars for error handling.
	var consecutiveErrors int
	lastErrorTime := time.Now()

	for {
		conn, err := listener.Accept()
		if err != nil {
			lastErrorTime, consecutiveErrors = acceptErrorHandle(err, lastErrorTime, consecutiveErrors)
			continue
		}
		
		go forwardServer(bindAddress, conn, domainName)
	}
	
}

/*
	gateway server. bridges connection from internet
	to client.
	- bindAddress: address to bind to
	- conn: tunnel connection
*/
func forwardServer(bindAddress string, conn net.Conn, domainName string) {
	var tag string = "[server-forwardServer] -"

	var stopChan chan int = make(chan int)

	// Adjust default config for yamux.Session.
	config := yamux.DefaultConfig()
	config.ConnectionWriteTimeout = 10 * time.Second
	muxConn, err := yamux.Client(conn, config)
	if err != nil {
		panic(err)
	}
	log.Printf("%s Accepted muxing session.\n", tag)

	listener, err := net.Listen("tcp", bindAddress + ":0")
	if err != nil {
		panic(err)
	}

	log.Printf("%s Forward-server listening on %s\n", tag, listener.Addr().String())

	// Send control info 
	// (address of gateway).
	listenerAddr := listener.Addr().String()
	controlMsg := "," + domainName + ":" + listenerAddr[strings.LastIndex(listenerAddr, ":") + 1:]
	controlMsg = padControlMessage(controlMsg)
	log.Printf("%s control msg: %s\n", tag, controlMsg)
	
	controlConn, err := muxConn.Open()
	if err != nil {
		panic(err)
	}

	// First of all, receive the shared
	// secret so we know this is a proper
	// client.
	secretBuf := make([]byte, 36)
	controlConn.Read(secretBuf)
	secretStr := string(secretBuf)
	if secretStr != "d12a1f29-065d-4d65-addf-fefa51ff019b" {
		log.Printf("%s Invalid secret. Fake client. Exiting this forwardServer().\n", tag)
		return
	}
	log.Printf("%s Authenticaed. Legitimate client.\n", tag)

	controlConn.Write([]byte(controlMsg))
	controlConn.Close()

	// Vars for error handling.
	var consecutiveErrors int
	lastErrorTime := time.Now()

	// Tunnel loop.
	for {
		conn, err := listener.Accept()
		if err != nil {
			lastErrorTime, consecutiveErrors = acceptErrorHandle(err, lastErrorTime, consecutiveErrors)
			continue
		}
		go copyConnection(muxConn, conn, stopChan)
		select {
			case <-stopChan:
				log.Printf("%s shutting down this forward server.\n", tag)
				return
			default:
		}
	}
}

/*
	The actual tunneling logic
	(copying from the gateway to the client)
	- muxConn: client connection
	- srcConn: user connecting to gateway
*/

func copyConnection(muxConn *yamux.Session, srcConn net.Conn, stopChan chan int) {
	var tag string = "[server-copyConnection] -"
	
	dstConn, err := openSessionWithRetry(muxConn)
	if err != nil {
		log.Printf("%s failed to open session from %s to %s.\nerror: %s\n", tag, srcConn.RemoteAddr().String(), muxConn.RemoteAddr().String(), err)
		stopChan <- 1
		return
	}
	defer dstConn.Close()
	defer srcConn.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(dstConn, srcConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(srcConn, dstConn)
		done <- struct{}{}
	}()

	<-done
} 

func Run(bindAddress string, controlPort int, domainName string) {
	muxServer(bindAddress, controlPort, domainName)
}

/*
	Pad the control message to be 80 bytes
	(control message must always be fixed 80 bytes).
*/
func padControlMessage(message string) string {
	return fmt.Sprintf("%080s", message)
}


func openSessionWithRetry(session *yamux.Session) (net.Conn, error) {
	var maxRetries int = 5
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err := session.Open()
		if err == nil {
			return conn, nil
		}

		if (err == yamux.ErrStreamsExhausted) || (err == yamux.ErrTimeout) {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("failed to open session after %d retries: %s", maxRetries, err)
}