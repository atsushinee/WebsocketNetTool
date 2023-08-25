

package websocket

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/klauspost/compress/flate"

	"nhooyr.io/websocket/internal/errd"
)



//

//


func (c *Conn) Writer(ctx context.Context, typ MessageType) (io.WriteCloser, error) {
	w, err := c.writer(ctx, typ)
	if err != nil {
		return nil, fmt.Errorf("failed to get writer: %w", err)
	}
	return w, nil
}


//

//


func (c *Conn) Write(ctx context.Context, typ MessageType, p []byte) error {
	_, err := c.write(ctx, typ, p)
	if err != nil {
		return fmt.Errorf("failed to write msg: %w", err)
	}
	return nil
}

type msgWriter struct {
	mw     *msgWriterState
	closed bool
}

func (mw *msgWriter) Write(p []byte) (int, error) {
	if mw.closed {
		return 0, errors.New("cannot use closed writer")
	}
	return mw.mw.Write(p)
}

func (mw *msgWriter) Close() error {
	if mw.closed {
		return errors.New("cannot use closed writer")
	}
	mw.closed = true
	return mw.mw.Close()
}

type msgWriterState struct {
	c *Conn

	mu      *mu
	writeMu *mu

	ctx    context.Context
	opcode opcode
	flate  bool

	trimWriter *trimLastFourBytesWriter
	dict       slidingWindow
}

func newMsgWriterState(c *Conn) *msgWriterState {
	mw := &msgWriterState{
		c:       c,
		mu:      newMu(c),
		writeMu: newMu(c),
	}
	return mw
}

func (mw *msgWriterState) ensureFlate() {
	if mw.trimWriter == nil {
		mw.trimWriter = &trimLastFourBytesWriter{
			w: writerFunc(mw.write),
		}
	}

	mw.dict.init(8192)
	mw.flate = true
}

func (mw *msgWriterState) flateContextTakeover() bool {
	if mw.c.client {
		return !mw.c.copts.clientNoContextTakeover
	}
	return !mw.c.copts.serverNoContextTakeover
}

func (c *Conn) writer(ctx context.Context, typ MessageType) (io.WriteCloser, error) {
	err := c.msgWriterState.reset(ctx, typ)
	if err != nil {
		return nil, err
	}
	return &msgWriter{
		mw:     c.msgWriterState,
		closed: false,
	}, nil
}

func (c *Conn) write(ctx context.Context, typ MessageType, p []byte) (int, error) {
	mw, err := c.writer(ctx, typ)
	if err != nil {
		return 0, err
	}

	if !c.flate() {
		defer c.msgWriterState.mu.unlock()
		return c.writeFrame(ctx, true, false, c.msgWriterState.opcode, p)
	}

	n, err := mw.Write(p)
	if err != nil {
		return n, err
	}

	err = mw.Close()
	return n, err
}

func (mw *msgWriterState) reset(ctx context.Context, typ MessageType) error {
	err := mw.mu.lock(ctx)
	if err != nil {
		return err
	}

	mw.ctx = ctx
	mw.opcode = opcode(typ)
	mw.flate = false

	mw.trimWriter.reset()

	return nil
}


func (mw *msgWriterState) Write(p []byte) (_ int, err error) {
	err = mw.writeMu.lock(mw.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to write: %w", err)
	}
	defer mw.writeMu.unlock()

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to write: %w", err)
			mw.c.close(err)
		}
	}()

	if mw.c.flate() {
		
		
		if mw.opcode != opContinuation && len(p) >= mw.c.flateThreshold {
			mw.ensureFlate()
		}
	}

	if mw.flate {
		err = flate.StatelessDeflate(mw.trimWriter, p, false, mw.dict.buf)
		if err != nil {
			return 0, err
		}
		mw.dict.write(p)
		return len(p), nil
	}

	return mw.write(p)
}

func (mw *msgWriterState) write(p []byte) (int, error) {
	n, err := mw.c.writeFrame(mw.ctx, false, mw.flate, mw.opcode, p)
	if err != nil {
		return n, fmt.Errorf("failed to write data frame: %w", err)
	}
	mw.opcode = opContinuation
	return n, nil
}


func (mw *msgWriterState) Close() (err error) {
	defer errd.Wrap(&err, "failed to close writer")

	err = mw.writeMu.lock(mw.ctx)
	if err != nil {
		return err
	}
	defer mw.writeMu.unlock()

	_, err = mw.c.writeFrame(mw.ctx, true, mw.flate, mw.opcode, nil)
	if err != nil {
		return fmt.Errorf("failed to write fin frame: %w", err)
	}

	if mw.flate && !mw.flateContextTakeover() {
		mw.dict.close()
	}
	mw.mu.unlock()
	return nil
}

func (mw *msgWriterState) close() {
	if mw.c.client {
		mw.c.writeFrameMu.forceLock()
		putBufioWriter(mw.c.bw)
	}

	mw.writeMu.forceLock()
	mw.dict.close()
}

func (c *Conn) writeControl(ctx context.Context, opcode opcode, p []byte) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	_, err := c.writeFrame(ctx, true, false, opcode, p)
	if err != nil {
		return fmt.Errorf("failed to write control frame %v: %w", opcode, err)
	}
	return nil
}


func (c *Conn) writeFrame(ctx context.Context, fin bool, flate bool, opcode opcode, p []byte) (_ int, err error) {
	err = c.writeFrameMu.lock(ctx)
	if err != nil {
		return 0, err
	}
	defer c.writeFrameMu.unlock()

	
	
	//
	
	
	c.closeMu.Lock()
	wroteClose := c.wroteClose
	c.closeMu.Unlock()
	if wroteClose && opcode != opClose {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-c.closed:
			return 0, c.closeErr
		}
	}

	select {
	case <-c.closed:
		return 0, c.closeErr
	case c.writeTimeout <- ctx:
	}

	defer func() {
		if err != nil {
			select {
			case <-c.closed:
				err = c.closeErr
			case <-ctx.Done():
				err = ctx.Err()
			}
			c.close(err)
			err = fmt.Errorf("failed to write frame: %w", err)
		}
	}()

	c.writeHeader.fin = fin
	c.writeHeader.opcode = opcode
	c.writeHeader.payloadLength = int64(len(p))

	if c.client {
		c.writeHeader.masked = true
		_, err = io.ReadFull(rand.Reader, c.writeHeaderBuf[:4])
		if err != nil {
			return 0, fmt.Errorf("failed to generate masking key: %w", err)
		}
		c.writeHeader.maskKey = binary.LittleEndian.Uint32(c.writeHeaderBuf[:])
	}

	c.writeHeader.rsv1 = false
	if flate && (opcode == opText || opcode == opBinary) {
		c.writeHeader.rsv1 = true
	}

	err = writeFrameHeader(c.writeHeader, c.bw, c.writeHeaderBuf[:])
	if err != nil {
		return 0, err
	}

	n, err := c.writeFramePayload(p)
	if err != nil {
		return n, err
	}

	if c.writeHeader.fin {
		err = c.bw.Flush()
		if err != nil {
			return n, fmt.Errorf("failed to flush: %w", err)
		}
	}

	select {
	case <-c.closed:
		return n, c.closeErr
	case c.writeTimeout <- context.Background():
	}

	return n, nil
}

func (c *Conn) writeFramePayload(p []byte) (n int, err error) {
	defer errd.Wrap(&err, "failed to write frame payload")

	if !c.writeHeader.masked {
		return c.bw.Write(p)
	}

	maskKey := c.writeHeader.maskKey
	for len(p) > 0 {
		
		if c.bw.Available() == 0 {
			err = c.bw.Flush()
			if err != nil {
				return n, err
			}
		}

		
		i := c.bw.Buffered()

		j := len(p)
		if j > c.bw.Available() {
			j = c.bw.Available()
		}

		_, err := c.bw.Write(p[:j])
		if err != nil {
			return n, err
		}

		maskKey = mask(maskKey, c.writeBuf[i:c.bw.Buffered()])

		p = p[j:]
		n += j
	}

	return n, nil
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}



func extractBufioWriterBuf(bw *bufio.Writer, w io.Writer) []byte {
	var writeBuf []byte
	bw.Reset(writerFunc(func(p2 []byte) (int, error) {
		writeBuf = p2[:cap(p2)]
		return len(p2), nil
	}))

	bw.WriteByte(0)
	bw.Flush()

	bw.Reset(w)

	return writeBuf
}

func (c *Conn) writeError(code StatusCode, err error) {
	c.setCloseErr(err)
	c.writeClose(code, err.Error())
	c.close(nil)
}
