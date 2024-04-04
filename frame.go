package errors

import (
	"fmt"
	"io"
	"strconv"
)

type Frame struct {
	Func string
	File string
	Line int
}

type ErrorTracer interface {
	ErrorTrace() []Frame
}

func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Func)
			io.WriteString(s, "\n  ")
			io.WriteString(s, f.File)
		default:
			io.WriteString(s, f.File)
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}
