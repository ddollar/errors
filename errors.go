package errors

import (
	stderrors "errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
)

var New = stderrors.New

var Errorf = fmt.Errorf

type wrappedError struct {
	error
	frame Frame
}

type ErrorTracer interface {
	ErrorTrace() []Frame
}

type Frame struct {
	Func string
	File string
	Line int
}

func Cause(err error) error {
	switch t := err.(type) {
	case wrappedError:
		return t.error
	default:
		return err
	}
}

func Wrap(err error) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}

	fnc := "unknown"

	if fn := runtime.FuncForPC(pc); fn != nil {
		fnc = fn.Name()
	}

	we := wrappedError{
		error: err,
		frame: Frame{
			Func: fnc,
			File: file,
			Line: line,
		},
	}

	return we
}

func (we wrappedError) Cause() error {
	return we.error
}

func (we wrappedError) ErrorTrace() []Frame {
	t := []Frame{we.frame}

	if et, ok := we.error.(ErrorTracer); ok {
		t = append(et.ErrorTrace(), t...)
	}

	return t
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
