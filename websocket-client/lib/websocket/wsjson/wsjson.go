
package wsjson 

import (
	"context"
	"encoding/json"
	"fmt"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/internal/bpool"
	"nhooyr.io/websocket/internal/errd"
)



func Read(ctx context.Context, c *websocket.Conn, v interface{}) error {
	return read(ctx, c, v)
}

func read(ctx context.Context, c *websocket.Conn, v interface{}) (err error) {
	defer errd.Wrap(&err, "failed to read JSON message")

	_, r, err := c.Reader(ctx)
	if err != nil {
		return err
	}

	b := bpool.Get()
	defer bpool.Put(b)

	_, err = b.ReadFrom(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b.Bytes(), v)
	if err != nil {
		c.Close(websocket.StatusInvalidFramePayloadData, "failed to unmarshal JSON")
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}



func Write(ctx context.Context, c *websocket.Conn, v interface{}) error {
	return write(ctx, c, v)
}

func write(ctx context.Context, c *websocket.Conn, v interface{}) (err error) {
	defer errd.Wrap(&err, "failed to write JSON message")

	w, err := c.Writer(ctx, websocket.MessageText)
	if err != nil {
		return err
	}

	
	
	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return w.Close()
}
