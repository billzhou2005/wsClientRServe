In this tutorial, we’ll be looking at how we can use the Gorilla Websocket package in Golang.

This library provides us with easy to write websocket client/servers in Go. It has a working production quality Go implementation of the websocket protocol, which enables us to deal with stateful http connections using websockets.

Let’s now understand how we can quickly set up a testable websocket application using Gorilla.

Installing Gorilla Websocket Go Package
There are no additional dependencies apart from a working Go compiler, so you just need to use go get!

1
go get github.com/gorilla/websocket
Websocket application Design
Before going on to any examples, let’s first design a rough layout of what needs to be done.

Any application using the websocket protocol generally needs a client and a server.

The server program binds to a port on the server and starts listening for any websocket connections. The connection related details are defined by the websocket protocol, which acts over a raw HTTP connection.

The client program tries to make a connection with the server using a websocket URL. Note that the client program does NOT need to be implemented using Golang, although Gorilla provides us with APIs for writing clients.

If you have a web application using a separate frontend, generally the websocket client would be implemented in that language (Javascript, etc)

However, for the purpose of illustration, we will be writing BOTH the client and the server programs in Go.

Now let’s get our client-server architecture running!
We’ll have a program server.go for the server, and client.go for the client.

Using Gorilla Websockets – Creating our server
The websocket server will be implemented over a regular http server. We’ll be using net/http for serving raw HTTP connections.

Now in server.go, let’s write our regular HTTP server and add a socketHandler() function to handle the websocket logic.

1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21
22
23
24
25
26
27
28
29
30
31
32
33
34
35
36
37
38
39
40
41
42
43
44
45
46
47
// server.go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func socketHandler(w http.ResponseWriter, r *http.Request) {
    // Upgrade our raw HTTP connection to a websocket based one
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("Error during connection upgradation:", err)
        return
    }
    defer conn.Close()

    // The event loop
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("Error during message reading:", err)
            break
        }
        log.Printf("Received: %s", message)
        err = conn.WriteMessage(messageType, message)
        if err != nil {
            log.Println("Error during message writing:", err)
            break
        }
    }
}

func home(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Index Page")
}

func main() {
    http.HandleFunc("/socket", socketHandler)
    http.HandleFunc("/", home)
    log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
The magic that gorilla does is to convert these raw HTTP connections into a stateful websocket connection, using a connection upgradation. This is why the library uses a struct called Upgrader to help us with that.

We use a global upgrader variable to help us convert any incoming HTTP connection into websocket protocol, via upgrader.Upgrade(). This will return to us a *websocket.Connection, which we can now use to deal with the websocket connection.

The server reads messages using conn.ReadMessage() and writes them back using conn.WriteMessage()

This server simply echoes any incoming websocket messages back to the client, so this shows how websockets can be used for full-duplex communication.

Now let’s move onto the client implementation at client.go.

Creating our client program
We’ll be writing the client also using Gorilla. This simple client will keep emitting messages after every 1 second. If our entire system works as intended, the server will receive packets spaced at an interval of 1 second and reply the same message back.

The client will also have functionality to receive incoming websocket packets. In our program, we will have a separate goroutine handler receiveHandler which listens for these incoming packets.
If you observe the code, you’ll notice that I created two channels done and interrupt for communication between receiveHandler() and main().

We use an infinite loop for listening to events through channels using select. We write a message using conn.WriteMessage() every second. If the interrupt signal is activated, any pending connections are closed and we exit gracefully!

The nested select is there to ensure two things:

If the receiveHandler channel exits, the channel 'done' will be closed. This is the first case <-done condition
If the 'done' channel does NOT close, there will be a timeout after 1 second, so the program WILL exit after the 1 second timeout
By carefully handling all cases using channels, select, you can have a minimal architecture which can be extended easily.

Let’s finally look at the output we get, on running both the client and the server!

Sample Output
Hopefully, you’ve got a good idea of how you could start writing your own WebSocket client/server programs in Golang!

This same style is used in a lot of open source projects in Github, so you could also poke around different repositories to get a hang of how this architecture works in actual practice. Until next time!

 // create connection
    // schema can be ws:// or wss://
    // host, port – WebSocket server
    conn, err := websocket.Dial("{schema}://{host}:{port}", "", op.Origin)
    if err != nil {
        // handle error
    }
    defer conn.Close()
             .......
      // send message
        if err = websocket.JSON.Send(conn, {message}); err != nil {
         // handle error
    }
              .......
        // receive message
    // messageType initializes some type of message
    message := messageType{}
    if err := websocket.JSON.Receive(conn, &message); err != nil {
          // handle error
    }
        .......