package csv

import (
	"bufio"
	"io"
	"strings"
)

type Reader interface {
	Read() ([]string, error)
}

func NewReader(r io.Reader, delimiter rune) *GoovernReader {
	return &GoovernReader{
		scanner:   bufio.NewScanner(r),
		delimiter: delimiter,
	}
}

type GoovernReader struct {
	scanner   *bufio.Scanner
	delimiter rune
}

func (r *GoovernReader) Read() ([]string, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := r.scanner.Text()
	return strings.Split(line, string(r.delimiter)), nil
}
