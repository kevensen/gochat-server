package main

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/rhtps/gochat/trace"
	"github.com/stretchr/objx"
)

type room struct {
	forward chan *message
	join    chan *client
	leave   chan *client
	clients map[*client]bool
	tracer  trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			glog.Infoln("New client joined -", client.userData["name"])
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			glog.Infoln("Client left -", client.userData["name"])
		case msg := <-r.forward:
			glog.Infoln("Message received -", msg.Message)
			for client := range r.clients {
				select {
				case client.send <- msg:
					glog.Infoln(" -- sent to client")
				default:
					delete(r.clients, client)
					close(client.send)
					glog.Infoln(" -- failed to send, cleaned up client")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		glog.Errorln("ServeHTTP:", err)
		return
	}
	authCookie, err := req.Cookie("auth")
	if err != nil {
		glog.Warningln("Failed to get auth cookie:", err)
		return
	}

	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize),
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()

}
