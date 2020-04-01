package net

import (
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"time"
)

const (
	PongWait       = 5 * time.Second
	MaxMessageSize = 1024
	WriteWait      = 5 * time.Second
	PingPeriod     = 100 * time.Millisecond
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
	conn.SetReadLimit(MaxMessageSize)
	conn.SetPongHandler(connection.pongHandler)
	go connection.writer()
	return connection
}

func (c *WebsocketConnection) pongHandler(_ string) error {
	return errors.Wrap(c.conn.SetReadDeadline(time.Now().Add(PongWait)), "setting read deadline after pong")
}

func (c *WebsocketConnection) writer() {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case message, ok := <-c.out:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.Close()
				return
			}

			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				c.Close()
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(WriteWait)); err != nil {
				c.Close()
				return
			}
		}
	}
}

func (c *WebsocketConnection) Write(message []byte) {
	select {
	case <-c.done:
	case c.out <- message:
	}
}

func (c *WebsocketConnection) Read() ([]byte, error) {
	_, message, err := c.conn.ReadMessage()
	return message, err
}

func (c *WebsocketConnection) Close() error {
	close(c.done)
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	return c.conn.Close()
}
