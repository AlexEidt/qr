package qr

import (
	"bytes"
	"fmt"
	"strconv"
)

type Buffer struct {
	size   int
	buffer bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		size:   0,
		buffer: bytes.Buffer{},
	}
}

func (b *Buffer) Size() int {
	return b.size
}

func (b *Buffer) String() string {
	return b.buffer.String()
}

func (b *Buffer) Bytes() []byte {
	size := b.size / 8
	if b.size%8 != 0 {
		size++
	}

	data := make([]byte, size)

	i := 0
	s := b.String()
	for ; i < len(s)-8; i += 8 {
		bits, _ := strconv.ParseUint(s[i:i+8], 2, 64)
		data[i/8] = byte(bits)
	}
	bits, _ := strconv.ParseUint(s[i:], 2, 64)
	data[i/8] = byte(bits)

	return data
}

func (b *Buffer) Add(val int, size int) {
	if size > 0 {
		b.buffer.WriteString(fmt.Sprintf("%0*b", size, val))
		b.size += size
	}
}

func (b *Buffer) Write(data string) {
	b.buffer.WriteString(data)
	b.size += len(data)
}

func (b *Buffer) Clear() {
	b.buffer.Reset()
	b.size = 0
}
