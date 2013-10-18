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
		socketUrl string
		replyText string
		verbose   bool
		delay     int
		nreplies  int
	)

	flag.StringVar(&socketUrl, "socket", "", "the mock server socket: ipc:///tmp/foo.sock, tcp://1.2.3.4:9999, etc")
	flag.StringVar(&replyText, "reply", "", "the text to return whenever any request is made")
	flag.BoolVar(&verbose, "verbose", false, "set true to log all input")
	flag.IntVar(&delay, "delay", 0, "the number of milliseconds to sleep before replying to requests")
	flag.IntVar(&nreplies, "n", 0, "the number of replies to return (default: 0 = unlimited)")
	flag.Parse()

	if socketUrl == "" {
		log.Println("main", "you must specify a socket")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if replyText == "" {
		log.Println("main", "you must specify reply text")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if nreplies < 0 {
		nreplies = 0
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

	err := startServerRep(socketUrl, replyText, verbose, delay, done, nreplies)
	if err != nil {
		log.Fatalln("startServer", err.Error())
		os.Exit(1)
	}
	<-done
}

type zmqMessage struct {
	Payload []byte
}

func startServerRep(socketUrl string, replyText string, verbose bool, delay int, done chan bool, nreplies int) (err error) {
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

	n := 0

	go func() {
		input := make(chan zmqMessage, 1)
		output := make(chan zmqMessage, 1)
		pollCh := zmqPoller.Poll()

		doneFunc := func() {
			zmqPoller.Close()
			zmqSocket.Close()
			zmqContext.Close()
			done <- true
		}

		for {
			select {
			case <-done:
				doneFunc()
				break

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

				output <- zmqMessage{Payload: []byte(replyText)}

			case outputMessage := <-output:
				time.Sleep(time.Duration(delay) * time.Millisecond)
				err := zmqSocket.Send(outputMessage.Payload, 0)
				if err != nil {
					log.Println("zmqSocket.Send", err.Error())
				}

				n++
				if nreplies > 0 && n >= nreplies {
					doneFunc()
					break
				}
			}
		}
	}()

	return
}
