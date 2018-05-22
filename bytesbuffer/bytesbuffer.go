package bytesbuffer

import (
	"errors"
	"io"
)

var ErrTooLarge = errors.New("bytes.Buffer: too large")

const maxInt = int(^uint(0) >> 1)

type Buffer struct {
	buf []byte

	off int
}

func (b *Buffer) empty() bool {
	return len(b.buf) <= b.off
}

func (b *Buffer) Len() int {
	return len(b.buf[b.off:])
}

func (b *Buffer) Cap() int {
	return cap(b.buf)
}

func (b *Buffer) Bytes() []byte {
	return b.buf[b.off:]
}

func (b *Buffer) String() string {
	if b == nil {
		return "<nil>"
	}

	return string(b.buf[b.off:])
}

func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

func (b *Buffer) Truncate(n int) {
	if n == 0 {
		b.Reset()
		return
	}
	if n < 0 || n > b.Len() {
		panic("bytes.Buffer: truncation out of range")
	}
	b.buf = b.buf[:b.off+n]
}

func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

func (b *Buffer) grow(n int) int {
	m := b.Len()

	if m == 0 && b.off != 0 {
		b.Reset()
	}

	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}

	c := cap(b.buf)
	if n <= c/2-m {
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}

	b.off = 0
	b.buf = b.buf[:m+n]

	return m
}

func makeSlice(n int) []byte {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

func (b *Buffer) WriteByte(c byte) error {
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = c
	return nil
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	m, ok := b.tryGrowByReslice(len(s))
	if !ok {
		m = b.grow(len(s))
	}
	return copy(b.buf[m:], s), nil
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.empty() {
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}

func (b *Buffer) ReadByte() (byte, error) {
	if b.empty() {
		b.Reset()
		return 0, io.EOF
	}
	c := b.buf[b.off]
	b.off++
	return c, nil
}

func NewBuffer(buf []byte) *Buffer {
	return &Buffer{
		buf: buf,
	}
}

func NewBufferString(s string) *Buffer {
	return &Buffer{
		buf: []byte(s),
	}
}
