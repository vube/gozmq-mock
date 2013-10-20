# gozmq-mock

*This is a work in progress, consider it "alpha" quality code. This is an implementation of a basic REP server that expects REQ requests and a basic REQ client that expects REP replies. Future versions will support DEALER/ROUTER.*

gozmq-mock provides a simple ZMQ client and server that sends static data either as a request or as a reply. The mock client and server both support artificial delays (using sleeps) and a request/response limit. The goal of the project is to create a client and server that can help you verify code is properly parsing the requests and responses and handling timeouts appropriately.

# Dependencies

* [alecthomas's gozmq](http://github.com/alecthomas/gozmq): `go get github.com/alecthomas/gozmq`
* [tchap's gozmq-poller](https://github.com/tchap/gozmq-poller): `go get github.com/tchap/gozmq-poller`

# Example usage

```bash
# REP server in one window
./gozmq-mock -socket="ipc:///tmp/mock.sock" -reply='{ "foo": "1" }' -verbose -n=5 -type=REP -delay=1000
# REQ client in another
./gozmq-mock -socket="ipc:///tmp/mock.sock" -reply='{ "bar": "2" }' -verbose -n=5 -type=REQ -delay=1000
```

These two mocks will talk to each other over the specified ZMQ socket. The output will look something like:

```
# REP server
2013/10/19 19:49:52 < request { "bar": "2" }
2013/10/19 19:49:53 > reply { "foo": "1" }
2013/10/19 19:49:53 < request { "bar": "2" }
2013/10/19 19:49:54 > reply { "foo": "1" }
2013/10/19 19:49:54 < request { "bar": "2" }
2013/10/19 19:49:55 > reply { "foo": "1" }
2013/10/19 19:49:55 < request { "bar": "2" }
2013/10/19 19:49:56 > reply { "foo": "1" }
2013/10/19 19:49:56 < request { "bar": "2" }
2013/10/19 19:49:57 > reply { "foo": "1" }

# REQ client
2013/10/19 19:49:01 > request { "bar": "2" }
2013/10/19 19:49:04 < reply { "foo": "1" }
2013/10/19 19:49:04 > request { "bar": "2" }
2013/10/19 19:49:07 < reply { "foo": "1" }
2013/10/19 19:49:07 > request { "bar": "2" }
2013/10/19 19:49:10 < reply { "foo": "1" }
2013/10/19 19:49:10 > request { "bar": "2" }
2013/10/19 19:49:13 < reply { "foo": "1" }
2013/10/19 19:49:13 > request { "bar": "2" }
2013/10/19 19:49:16 < reply { "foo": "1" }
```

# License

The MIT License (MIT)

Copyright (c) 2013 The Vubeologists

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
