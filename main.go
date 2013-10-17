// A gozmq server that responds to requests in a specified manner,
// allowing you to mock ZMQ servers.

package main

import (
	"flag"
	zmq "github.com/alecthomas/gozmq"
	poller "github.com/tchap/gozmq-poller"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var (
		socketUrl    string
		responseText string
		verbose      bool
		delay        int
	)

	flag.StringVar(&socketUrl, "socket", "", "the mock server socket: ipc:///tmp/foo.sock, tcp://1.2.3.4:9999, etc")
	flag.StringVar(&responseText, "response", "", "the text to return whenever any request is made")
	flag.BoolVar(&verbose, "verbose", false, "set true to log all input")
	flag.IntVar(&delay, "delay", 0, "the number of milliseconds to sleep before replying to requests")
	flag.Parse()

	if socketUrl == "" {
		log.Println("main", "you must specify a socket")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if responseText == "" {
		log.Println("main", "you must specify response text")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// OK, we're all set. Get ready to start the server and to wait for
	// a signal to end it.

	done := make(chan bool, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("main", "caught signal, going down")
		done <- true
		<-done
		os.Exit(0)
	}()

	err := startServerRep(socketUrl, responseText, verbose, delay, done)
	if err != nil {
		log.Fatalln("startServer", err.Error())
		os.Exit(1)
	}
}

type zmqMessage struct {
	Payload []byte
}

func startServerRep(socketUrl string, responseText string, verbose bool, delay int, done chan bool) (err error) {
	zmqContext, err := zmq.NewContext()
	if err != nil {
		return
	}

	zmqSocket, err := zmqContext.NewSocket(zmq.REP)
	if err != nil {
		return
	}

	err = zmqSocket.Bind(socketUrl)
	if err != nil {
		return
	}

	pf := poller.NewFactory(zmqContext)

	zmqPoller, err := pf.NewPoller(zmq.PollItems{
		0: zmq.PollItem{Socket: zmqSocket, Events: zmq.POLLIN},
	})
	if err != nil {
		return
	}

	go func() {
		input := make(chan zmqMessage, 1)
		output := make(chan zmqMessage, 1)
		pollCh := zmqPoller.Poll()

		for {
			select {
			case <-done:
				zmqSocket.Close()
				zmqContext.Close()
				zmqPoller.Close()
				done <- true
				return

			case <-pollCh:
				msg, _ := zmqSocket.Recv(0)

				input <- zmqMessage{Payload: msg}
				// Reset the poller so it can accept more requests
				pollCh = zmqPoller.Poll()

			case inputMessage := <-input:
				if verbose {
					log.Println("input", string(inputMessage.Payload))
				}

				// Read the request from the caller, reply to the request
				// as expected.

				output <- zmqMessage{Payload: []byte(responseText)}

			case outputMessage := <-output:
				time.Sleep(time.Duration(delay) * time.Millisecond)
				err := zmqSocket.Send(outputMessage.Payload, 0)
				if err != nil {
					log.Println("zmqSocket.Send", err.Error())
				}
			}
		}
	}()

	select {}

	return
}
