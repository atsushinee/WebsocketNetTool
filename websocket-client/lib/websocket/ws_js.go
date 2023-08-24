package websocket 

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall/js"

	"nhooyr.io/websocket/internal/bpool"
	"nhooyr.io/websocket/internal/wsjs"
	"nhooyr.io/websocket/internal/xsync"
)


type Conn struct {
	ws wsjs.WebSocket

	
	msgReadLimit xsync.Int64

	closingMu     sync.Mutex
	isReadClosed  xsync.Int64
	closeOnce     sync.Once
	closed        chan struct{}
	closeErrOnce  sync.Once
	closeErr      error
	closeWasClean bool

	releaseOnClose   func()
	releaseOnMessage func()

	readSignal chan struct{}
	readBufMu  sync.Mutex
	readBuf    []wsjs.MessageEvent
}

func (c *Conn) close(err error, wasClean bool) {
	c.closeOnce.Do(func() {
		runtime.SetFinalizer(c, nil)

		if !wasClean {
			err = fmt.Errorf("unclean connection close: %w", err)
		}
		c.setCloseErr(err)
		c.closeWasClean = wasClean
		close(c.closed)
	})
}

func (c *Conn) init() {
	c.closed = make(chan struct{})
	c.readSignal = make(chan struct{}, 1)

	c.msgReadLimit.Store(32768)

	c.releaseOnClose = c.ws.OnClose(func(e wsjs.CloseEvent) {
		err := CloseError{
			Code:   StatusCode(e.Code),
			Reason: e.Reason,
		}
		
		
		
		c.close(err, e.WasClean)

		c.releaseOnClose()
		c.releaseOnMessage()
	})

	c.releaseOnMessage = c.ws.OnMessage(func(e wsjs.MessageEvent) {
		c.readBufMu.Lock()
		defer c.readBufMu.Unlock()

		c.readBuf = append(c.readBuf, e)

		
		select {
		case c.readSignal <- struct{}{}:
		default:
		}
	})

	runtime.SetFinalizer(c, func(c *Conn) {
		c.setCloseErr(errors.New("connection garbage collected"))
		c.closeWithInternal()
	})
}

func (c *Conn) closeWithInternal() {
	c.Close(StatusInternalError, "something went wrong")
}



func (c *Conn) Read(ctx context.Context) (MessageType, []byte, error) {
	if c.isReadClosed.Load() == 1 {
		return 0, nil, errors.New("WebSocket connection read closed")
	}

	typ, p, err := c.read(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read: %w", err)
	}
	if int64(len(p)) > c.msgReadLimit.Load() {
		err := fmt.Errorf("read limited at %v bytes", c.msgReadLimit.Load())
		c.Close(StatusMessageTooBig, err.Error())
		return 0, nil, err
	}
	return typ, p, nil
}

func (c *Conn) read(ctx context.Context) (MessageType, []byte, error) {
	select {
	case <-ctx.Done():
		c.Close(StatusPolicyViolation, "read timed out")
		return 0, nil, ctx.Err()
	case <-c.readSignal:
	case <-c.closed:
		return 0, nil, c.closeErr
	}

	c.readBufMu.Lock()
	defer c.readBufMu.Unlock()

	me := c.readBuf[0]
	
	
	copy(c.readBuf, c.readBuf[1:])
	c.readBuf = c.readBuf[:len(c.readBuf)-1]

	if len(c.readBuf) > 0 {
		
		select {
		case c.readSignal <- struct{}{}:
		default:
		}
	}

	switch p := me.Data.(type) {
	case string:
		return MessageText, []byte(p), nil
	case []byte:
		return MessageBinary, p, nil
	default:
		panic("websocket: unexpected data type from wsjs OnMessage: " + reflect.TypeOf(me.Data).String())
	}
}


func (c *Conn) Ping(ctx context.Context) error {
	return nil
}



func (c *Conn) Write(ctx context.Context, typ MessageType, p []byte) error {
	err := c.write(ctx, typ, p)
	if err != nil {
		
		
		
		
		err := fmt.Errorf("failed to write: %w", err)
		c.setCloseErr(err)
		c.closeWithInternal()
		return err
	}
	return nil
}

func (c *Conn) write(ctx context.Context, typ MessageType, p []byte) error {
	if c.isClosed() {
		return c.closeErr
	}
	switch typ {
	case MessageBinary:
		return c.ws.SendBytes(p)
	case MessageText:
		return c.ws.SendText(string(p))
	default:
		return fmt.Errorf("unexpected message type: %v", typ)
	}
}





func (c *Conn) Close(code StatusCode, reason string) error {
	err := c.exportedClose(code, reason)
	if err != nil {
		return fmt.Errorf("failed to close WebSocket: %w", err)
	}
	return nil
}

func (c *Conn) exportedClose(code StatusCode, reason string) error {
	c.closingMu.Lock()
	defer c.closingMu.Unlock()

	ce := fmt.Errorf("sent close: %w", CloseError{
		Code:   code,
		Reason: reason,
	})

	if c.isClosed() {
		return fmt.Errorf("tried to close with %q but connection already closed: %w", ce, c.closeErr)
	}

	c.setCloseErr(ce)
	err := c.ws.Close(int(code), reason)
	if err != nil {
		return err
	}

	<-c.closed
	if !c.closeWasClean {
		return c.closeErr
	}
	return nil
}



func (c *Conn) Subprotocol() string {
	return c.ws.Subprotocol()
}


type DialOptions struct {
	
	Subprotocols []string
}





func Dial(ctx context.Context, url string, opts *DialOptions) (*Conn, *http.Response, error) {
	c, resp, err := dial(ctx, url, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to WebSocket dial %q: %w", url, err)
	}
	return c, resp, nil
}

func dial(ctx context.Context, url string, opts *DialOptions) (*Conn, *http.Response, error) {
	if opts == nil {
		opts = &DialOptions{}
	}

	url = strings.Replace(url, "http://", "ws://", 1)
	url = strings.Replace(url, "https://", "wss://", 1)

	ws, err := wsjs.New(url, opts.Subprotocols)
	if err != nil {
		return nil, nil, err
	}

	c := &Conn{
		ws: ws,
	}
	c.init()

	opench := make(chan struct{})
	releaseOpen := ws.OnOpen(func(e js.Value) {
		close(opench)
	})
	defer releaseOpen()

	select {
	case <-ctx.Done():
		c.Close(StatusPolicyViolation, "dial timed out")
		return nil, nil, ctx.Err()
	case <-opench:
		return c, &http.Response{
			StatusCode: http.StatusSwitchingProtocols,
		}, nil
	case <-c.closed:
		return nil, nil, c.closeErr
	}
}



func (c *Conn) Reader(ctx context.Context) (MessageType, io.Reader, error) {
	typ, p, err := c.Read(ctx)
	if err != nil {
		return 0, nil, err
	}
	return typ, bytes.NewReader(p), nil
}




func (c *Conn) Writer(ctx context.Context, typ MessageType) (io.WriteCloser, error) {
	return writer{
		c:   c,
		ctx: ctx,
		typ: typ,
		b:   bpool.Get(),
	}, nil
}

type writer struct {
	closed bool

	c   *Conn
	ctx context.Context
	typ MessageType

	b *bytes.Buffer
}

func (w writer) Write(p []byte) (int, error) {
	if w.closed {
		return 0, errors.New("cannot write to closed writer")
	}
	n, err := w.b.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed to write message: %w", err)
	}
	return n, nil
}

func (w writer) Close() error {
	if w.closed {
		return errors.New("cannot close closed writer")
	}
	w.closed = true
	defer bpool.Put(w.b)

	err := w.c.Write(w.ctx, w.typ, w.b.Bytes())
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}


func (c *Conn) CloseRead(ctx context.Context) context.Context {
	c.isReadClosed.Store(1)

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		c.read(ctx)
		c.Close(StatusPolicyViolation, "unexpected data message")
	}()
	return ctx
}


func (c *Conn) SetReadLimit(n int64) {
	c.msgReadLimit.Store(n)
}

func (c *Conn) setCloseErr(err error) {
	c.closeErrOnce.Do(func() {
		c.closeErr = fmt.Errorf("WebSocket closed: %w", err)
	})
}

func (c *Conn) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}
