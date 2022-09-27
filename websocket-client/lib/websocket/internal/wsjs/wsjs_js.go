




package wsjs

import (
	"syscall/js"
)

func handleJSError(err *error, onErr func()) {
	r := recover()

	if jsErr, ok := r.(js.Error); ok {
		*err = jsErr

		if onErr != nil {
			onErr()
		}
		return
	}

	if r != nil {
		panic(r)
	}
}


func New(url string, protocols []string) (c WebSocket, err error) {
	defer handleJSError(&err, func() {
		c = WebSocket{}
	})

	jsProtocols := make([]interface{}, len(protocols))
	for i, p := range protocols {
		jsProtocols[i] = p
	}

	c = WebSocket{
		v: js.Global().Get("WebSocket").New(url, jsProtocols),
	}

	c.setBinaryType("arraybuffer")

	return c, nil
}


type WebSocket struct {
	v js.Value
}

func (c WebSocket) setBinaryType(typ string) {
	c.v.Set("binaryType", string(typ))
}

func (c WebSocket) addEventListener(eventType string, fn func(e js.Value)) func() {
	f := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fn(args[0])
		return nil
	})
	c.v.Call("addEventListener", eventType, f)

	return func() {
		c.v.Call("removeEventListener", eventType, f)
		f.Release()
	}
}


type CloseEvent struct {
	Code     uint16
	Reason   string
	WasClean bool
}


func (c WebSocket) OnClose(fn func(CloseEvent)) (remove func()) {
	return c.addEventListener("close", func(e js.Value) {
		ce := CloseEvent{
			Code:     uint16(e.Get("code").Int()),
			Reason:   e.Get("reason").String(),
			WasClean: e.Get("wasClean").Bool(),
		}
		fn(ce)
	})
}



func (c WebSocket) OnError(fn func(e js.Value)) (remove func()) {
	return c.addEventListener("error", fn)
}


type MessageEvent struct {
	
	Data interface{}

	
	
}


func (c WebSocket) OnMessage(fn func(m MessageEvent)) (remove func()) {
	return c.addEventListener("message", func(e js.Value) {
		var data interface{}

		arrayBuffer := e.Get("data")
		if arrayBuffer.Type() == js.TypeString {
			data = arrayBuffer.String()
		} else {
			data = extractArrayBuffer(arrayBuffer)
		}

		me := MessageEvent{
			Data: data,
		}
		fn(me)

		return
	})
}


func (c WebSocket) Subprotocol() string {
	return c.v.Get("protocol").String()
}


func (c WebSocket) OnOpen(fn func(e js.Value)) (remove func()) {
	return c.addEventListener("open", fn)
}


func (c WebSocket) Close(code int, reason string) (err error) {
	defer handleJSError(&err, nil)
	c.v.Call("close", code, reason)
	return err
}



func (c WebSocket) SendText(v string) (err error) {
	defer handleJSError(&err, nil)
	c.v.Call("send", v)
	return err
}



func (c WebSocket) SendBytes(v []byte) (err error) {
	defer handleJSError(&err, nil)
	c.v.Call("send", uint8Array(v))
	return err
}

func extractArrayBuffer(arrayBuffer js.Value) []byte {
	uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
	dst := make([]byte, uint8Array.Length())
	js.CopyBytesToGo(dst, uint8Array)
	return dst
}

func uint8Array(src []byte) js.Value {
	uint8Array := js.Global().Get("Uint8Array").New(len(src))
	js.CopyBytesToJS(uint8Array, src)
	return uint8Array
}
