package main

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type room struct {
	forward chan *message
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			glog.Infoln("Room - new client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			glog.Infoln("Room - client left")
		case msg := <-r.forward:
			glog.Infoln("Room - message received -", msg.Message)
			for client := range r.clients {
				select {
				case client.send <- msg:
					glog.Infoln(msg, " -- sent to client")
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

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		glog.Errorln("ServeHTTP:", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan *message, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()

}
