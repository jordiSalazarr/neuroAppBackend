package utils

import (
	"fmt"
	"io"
)

func ReadAllLimited(r io.Reader, max int64) ([]byte, error) {
	if max <= 0 {
		return io.ReadAll(r)
	}
	lr := io.LimitedReader{R: r, N: max}
	b, err := io.ReadAll(&lr)
	if err != nil {
		return nil, err
	}
	if lr.N == 0 {
		return nil, fmt.Errorf("payload exceeds limit (%d bytes)", max)
	}
	return b, nil
}
