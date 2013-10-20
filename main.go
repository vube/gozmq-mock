// A gozmq server that either responds to requests (REP) with static
// data or makes repeated requests (REQ) with static data to a
// specified ZMQ server.

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
		socketUrl  string
		socketType string
		replyText  string
		verbose    bool
		delay      int
		nreplies   int
	)

	flag.StringVar(&socketUrl, "socket", "", "the mock server socket: ipc:///tmp/foo.sock, tcp://1.2.3.4:9999, etc")
	flag.StringVar(&socketType, "type", "REP", "the type of socket (currently supported: REP (default), REQ)")
	flag.StringVar(&replyText, "reply", "", "the text to return whenever any request is made")
	flag.BoolVar(&verbose, "verbose", false, "set true to log all input (REP) or output (REQ)")
	flag.IntVar(&delay, "delay", 0, "the number of milliseconds to sleep between messages")
	flag.IntVar(&nreplies, "n", 0, "the number of replies or requests to send (default: 0 = unlimited)")
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
	finishedClosing := make(chan bool, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("main", "caught signal, going down")
		done <- true
		<-done
		os.Exit(0)
	}()

	var err error
	switch socketType {
	case "REP":
		err = startServerRep(socketUrl, replyText, verbose, delay, done, finishedClosing, nreplies)
	case "REQ":
		err = startServerReq(socketUrl, replyText, verbose, delay, done, finishedClosing, nreplies)
	default:
		log.Println("main", "only supports REP and REQ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalln("startServer", err.Error())
		os.Exit(1)
	}
	<-finishedClosing
}

type zmqMessage struct {
	Payload []byte
}

// Start a simple "request" sender. This just sends the same request
// payload repeatedly to a server at the specified socketUrl until
// a value is passed to the done channel or the number of requests
// exceeds nrequests. (nrequests = 0 means unlimited requests).
func startServerReq(socketUrl string, payload string, verbose bool, delay int, done chan bool, finishedClosing chan bool, nrequests int) (err error) {
	zmqContext, err := zmq.NewContext()
	if err != nil {
		return
	}

	zmqSocket, err := zmqContext.NewSocket(zmq.REQ)
	if err != nil {
		return
	}

	err = zmqSocket.Connect(socketUrl)
	if err != nil {
		return
	}

	go func() {
		n := 0

		for {
			select {
			case <-done:
				zmqSocket.Close()
				zmqContext.Close()
				finishedClosing <- true
				return

			default:
				err := zmqSocket.Send([]byte(payload), 0)
				if err != nil {
					log.Println("zmqSocket.Send", err.Error())
					done <- true
					break
				}

				if verbose {
					log.Println("> request", string(payload))
				}

				if delay > 0 {
					time.Sleep(time.Duration(delay) * time.Millisecond)
				}

				reply, err := zmqSocket.Recv(0)
				if err != nil {
					log.Println("zmqSocket.Recv", err.Error())
					done <- true
					break
				}

				if verbose {
					log.Println("< reply", string(reply))
				}

				n++
				if nrequests > 0 && n >= nrequests {
					done <- true
				}
			}
		}
	}()

	return
}

// Start a simple "reply" server. This responds to every incoming
// request at the specified socket with the reply payload until
// a value is passed to the done channel or there have been at least
// nreplies sent.
func startServerRep(socketUrl string, replyText string, verbose bool, delay int, done chan bool, finishedClosing chan bool, nreplies int) (err error) {
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
		pollCh := zmqPoller.Poll()

		for {
			select {
			case <-done:
				zmqPoller.Close()
				zmqSocket.Close()
				zmqContext.Close()
				finishedClosing <- true
				return

			case <-pollCh:
				inputPayload, err := zmqSocket.Recv(0)
				if err != nil {
					log.Println("zmqSocket.Recv", err.Error())
					done <- true
					break
				}

				if verbose {
					log.Println("< request", string(inputPayload))
				}

				if delay > 0 {
					time.Sleep(time.Duration(delay) * time.Millisecond)
				}

				zmqSocket.Send([]byte(replyText), 0)

				if verbose {
					log.Println("> reply", string(replyText))
				}

				n++
				if nreplies > 0 && n >= nreplies {
					done <- true
				}

				pollCh = zmqPoller.Poll()
			}
		}
	}()

	return
}
