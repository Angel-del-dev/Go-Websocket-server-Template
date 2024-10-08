package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	/*
		When a client joins the websocket, this function will be called;
		Maps are not concurrent safe(Use a mutex instead);
	*/
	s.conns[ws] = true
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	/*
		1024 is the size to allocate;
		Might have to tweak the value in more complex processes;
	*/
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				/*
					If io.EOF is triggered, it means a client has
					lost or abandoned the connection.
					There's multiple aproaches, but in this case it will
					be deleted from Server.conns and drop
					the connection exiting the loop;
				*/
				delete(s.conns, ws)
				break
			}
			/*
				Other types of errors might be pressent as well;
			*/
			fmt.Println("read error:", err)
			continue
		}

		/*
			In this case, a client has sent some bytes to the server
			The message can be spread to the []byte slice into the msg variable;
		*/
		msg := buf[:n]
		/*
			Handle whatever it needs to be done with the msg variable;
			In this case its broadcasted to everyone in the Server.conns map
		*/
		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		/*
			go func() {}() creates an asyncronous and anonymous function
		*/
		go func(ws *websocket.Conn) {
			/*
				Basic if statement that creates a variable and checks its value in the same condition
			*/
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error:", err)
			}
		}(ws)
	}
}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.ListenAndServe(":3000", nil)
}
