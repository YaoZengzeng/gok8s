package pipe

import (
	"io"
	"sync"
)

type pipe struct {
	wc   chan []byte
	rc   chan int
	done chan struct{}

	sync.Mutex

	once sync.Once
	werr error
	rerr error
}

func (p *pipe) read(data []byte) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.readError()
	default:
	}

	select {
	case <-p.done:
		return 0, p.readError()
	case d := <-p.wc:
		n := copy(data, d)
		p.rc <- n
		return n, nil
	}
}

func (p *pipe) readError() error {
	if p.werr != nil {
		return p.werr
	}
	return io.ErrClosedPipe
}

func (p *pipe) write(data []byte) (n int, err error) {
	select {
	case <-p.done:
		return 0, p.writeError()
	default:
		p.Lock()
		defer p.Unlock()
	}

	for once := true; once || len(data) > 0; once = false {
		select {
		case p.wc <- data:
			nw := <-p.rc
			n += nw
			data = data[nw:]
		case <-p.done:
			return n, p.writeError()
		}
	}
	return
}

func (p *pipe) writeError() error {
	if p.rerr != nil {
		return p.rerr
	}
	return io.ErrClosedPipe
}

func (p *pipe) writeCloseWithError(err error) error {
	if err == nil {
		p.werr = io.EOF
	} else {
		p.werr = err
	}
	p.once.Do(func() { close(p.done) })
	return nil
}

func (p *pipe) readCloseWithError(err error) error {
	if err == nil {
		p.rerr = io.ErrClosedPipe
	} else {
		p.rerr = err
	}
	p.once.Do(func() { close(p.done) })
	return nil
}

type PipeWriter struct {
	p *pipe
}

func (w *PipeWriter) Write(data []byte) (n int, err error) {
	return w.p.write(data)
}

func (w *PipeWriter) Close() error {
	return w.p.writeCloseWithError(nil)
}

func (w *PipeWriter) CloseWithError(err error) error {
	return w.p.writeCloseWithError(err)
}

type PipeReader struct {
	p *pipe
}

func (r *PipeReader) Read(data []byte) (n int, err error) {
	return r.p.read(data)
}

func (r *PipeReader) Close() error {
	return r.p.readCloseWithError(nil)
}

func (r *PipeReader) CloseWithError(err error) error {
	return r.p.readCloseWithError(err)
}

func Pipe() (*PipeReader, *PipeWriter) {
	p := &pipe{
		wc:   make(chan []byte),
		rc:   make(chan int),
		done: make(chan struct{}),
	}

	return &PipeReader{p}, &PipeWriter{p}
}
