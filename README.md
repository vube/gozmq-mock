# gozmq-mock

*This is a work in progress, consider it "alpha" quality code. This is an implementation of a basic REP listener that expects REQ inputs. Future versions will support DEALER/ROUTER.*

gozmq-mock provides a simple ZMQ server that replies to all requests with a specified string pattern after waiting a configurable amount of time. This can be used to ensure that your client code is properly parsing the server reply and handling timeouts appropriately.

# Dependencies

[alecthomas's gozmq](http://github.com/alecthomas/gozmq): `go get github.com/alecthomas/gozmq`
[tchap's gozmq-poller](https://github.com/tchap/gozmq-poller): `go get github.com/tchap/gozmq-poller`

# Example usage

<code>
./gozmq-mock -socket="ipc:///tmp/mock.sock" -reply='{ "foo": "1" }' -verbose -n=123
</code>

Any zmq REQ packets sent to the mock will receive the specified reply. After the 123rd reply is sent the server will shut down.

## Example REQ client

<code>
package main

import "fmt"
import "github.com/alecthomas/gozmq"

func main() {
	zmqContext, _ := gozmq.NewContext()
	zmqSocket, _ := zmqContext.NewSocket(gozmq.REQ)
	zmqSocket.Connect("ipc:///tmp/mock.sock")

	zmqSocket.Send([]byte("foo"), 0)
	reply, _ := zmqSocket.Recv(0)

	// This will print whatever the mock returned (e.g. '{ "foo": "1" }')
	fmt.Println("reply:", string(reply))
}
</code>