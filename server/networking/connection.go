package networking

import (
	"github.com/gorilla/websocket"
)

// Connection provides a simple, generalized interface for networked connections in which there can be one reader and many writers.
// All connection errors are propagated through subsequent calls; that is, if any method invocation returns an error, all subsequent invocations of any method will also return the same error.
type Connection interface {
	// Write writes the contents of a byte array to the connection. It is safe for multiple threads to call Write simultaneously. Write blocks until either the message is sent or the connection is closed.
	Write([]byte)
	// Read reads the next message from a connection. It will block until either a message becomes available or the connection is closed.
	Read() ([]byte, error)
	// Close terminates a connection.
	Close() error
}

type WebsocketConnection struct {
	conn *websocket.Conn
	// done is a signal channel to indicate when the connection is closed
	// it is only called from the writer
	done chan struct{}
	out  chan []byte
}

func Websocket(conn *websocket.Conn) Connection {
	connection := &WebsocketConnection{
		conn: conn,
		done: make(chan struct{}),
		out:  make(chan []byte),
	}
	go connection.writer()
	return connection
}

func (c *WebsocketConnection) writer() {
	for message := range c.out {
		if c.conn.WriteMessage(websocket.BinaryMessage, message) != nil {
			close(c.done)
		}
	}
}

func (c *WebsocketConnection) Write(message []byte) {
	for {
		select {
		case <-c.done:
			return
		case c.out <- message:
			return
		}
	}
}

func (c *WebsocketConnection) Read() ([]byte, error) {
	_, message, err := c.conn.ReadMessage()
	return message, err
}

func (c *WebsocketConnection) Close() error {
	return c.conn.Close()
}
