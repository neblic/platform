package internal

import (
	"fmt"
	"io"
)

type Writer struct {
	io.Writer
}

func NewFormattedWriter(writer io.Writer) *Writer {
	return &Writer{
		Writer: writer,
	}
}

func (w *Writer) WriteString(str string) {
	w.Writer.Write([]byte(str))
}

func (w *Writer) WriteStringf(format string, a ...any) {
	w.Writer.Write([]byte(fmt.Sprintf(format, a...)))
}
