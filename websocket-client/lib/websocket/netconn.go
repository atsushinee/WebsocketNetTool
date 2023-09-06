package websocket

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"
)


//




//


//


//


//

//



//


//


func NetConn(ctx context.Context, c *Conn, msgType MessageType) net.Conn {
	nc := &netConn{
		c:       c,
		msgType: msgType,
	}

	var cancel context.CancelFunc
	nc.writeContext, cancel = context.WithCancel(ctx)
	nc.writeTimer = time.AfterFunc(math.MaxInt64, cancel)
	if !nc.writeTimer.Stop() {
		<-nc.writeTimer.C
	}

	nc.readContext, cancel = context.WithCancel(ctx)
	nc.readTimer = time.AfterFunc(math.MaxInt64, cancel)
	if !nc.readTimer.Stop() {
		<-nc.readTimer.C
	}

	return nc
}

type netConn struct {
	c       *Conn
	msgType MessageType

	writeTimer   *time.Timer
	writeContext context.Context

	readTimer   *time.Timer
	readContext context.Context

	readMu sync.Mutex
	eofed  bool
	reader io.Reader
}

var _ net.Conn = &netConn{}

func (c *netConn) Close() error {
	return c.c.Close(StatusNormalClosure, "")
}

func (c *netConn) Write(p []byte) (int, error) {
	err := c.c.Write(c.writeContext, c.msgType, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *netConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if c.eofed {
		return 0, io.EOF
	}

	if c.reader == nil {
		typ, r, err := c.c.Reader(c.readContext)
		if err != nil {
			switch CloseStatus(err) {
			case StatusNormalClosure, StatusGoingAway:
				c.eofed = true
				return 0, io.EOF
			}
			return 0, err
		}
		if typ != c.msgType {
			err := fmt.Errorf("unexpected frame type read (expected %v): %v", c.msgType, typ)
			c.c.Close(StatusUnsupportedData, err.Error())
			return 0, err
		}
		c.reader = r
	}

	n, err := c.reader.Read(p)
	if err == io.EOF {
		c.reader = nil
		err = nil
	}
	return n, err
}

type websocketAddr struct {
}

func (a websocketAddr) Network() string {
	return "websocket"
}

func (a websocketAddr) String() string {
	return "websocket/unknown-addr"
}

func (c *netConn) RemoteAddr() net.Addr {
	return websocketAddr{}
}

func (c *netConn) LocalAddr() net.Addr {
	return websocketAddr{}
}

func (c *netConn) SetDeadline(t time.Time) error {
	c.SetWriteDeadline(t)
	c.SetReadDeadline(t)
	return nil
}

func (c *netConn) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		c.writeTimer.Stop()
	} else {
		c.writeTimer.Reset(t.Sub(time.Now()))
	}
	return nil
}

func (c *netConn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		c.readTimer.Stop()
	} else {
		c.readTimer.Reset(t.Sub(time.Now()))
	}
	return nil
}
