package errors

import (
	stderrors "errors"
	"fmt"
	"runtime"
)

func Cause(err error) error {
	switch t := err.(type) {
	case wrappedError:
		return t.error
	case *wrappedError:
		return t.error
	default:
		return err
	}
}

func Errorf(format string, args ...interface{}) error {
	return Wrap(fmt.Errorf(format, args...))
}

func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

func Join(errs ...error) error {
	wrapped := make([]error, len(errs))

	for i := range errs {
		wrapped[i] = Wrap(errs[i])
	}

	return stderrors.Join(wrapped...)
}

func New(text string) error {
	return Wrap(stderrors.New(text))
}

func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}

	fnc := "unknown"

	if fn := runtime.FuncForPC(pc); fn != nil {
		fnc = fn.Name()
	}

	we := &wrappedError{
		error: err,
		frame: Frame{
			Func: fnc,
			File: file,
			Line: line,
		},
	}

	return we
}
