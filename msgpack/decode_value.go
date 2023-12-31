package msgpack

import (
	"fmt"
	"io"
)

type unexpectedCodeError struct {
	code byte
	hint string
}

func (err unexpectedCodeError) Error() string {
	return fmt.Sprintf("msgpack: unexpected code=%x decoding %s", err.code, err.hint)
}

func readN(r io.Reader, b []byte, n int) ([]byte, error) {
	if b == nil {
		if n == 0 {
			return make([]byte, 0), nil
		}
		switch {
		case n < 64:
			b = make([]byte, 0, 64)
		case n <= bytesAllocLimit:
			b = make([]byte, 0, n)
		default:
			b = make([]byte, 0, bytesAllocLimit)
		}
	}

	if n <= cap(b) {
		b = b[:n]
		_, err := io.ReadFull(r, b)
		return b, err
	}
	b = b[:cap(b)]

	var pos int
	for {
		alloc := min(n-len(b), bytesAllocLimit)
		b = append(b, make([]byte, alloc)...)

		_, err := io.ReadFull(r, b[pos:])
		if err != nil {
			return b, err
		}

		if len(b) == n {
			break
		}
		pos = len(b)
	}

	return b, nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
