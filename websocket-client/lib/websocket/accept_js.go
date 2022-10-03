package websocket

import (
	"errors"
	"net/http"
)


type AcceptOptions struct {
	Subprotocols         []string
	InsecureSkipVerify   bool
	OriginPatterns       []string
	CompressionMode      CompressionMode
	CompressionThreshold int
}


func Accept(w http.ResponseWriter, r *http.Request, opts *AcceptOptions) (*Conn, error) {
	return nil, errors.New("unimplemented")
}
