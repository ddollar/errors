package errors

type wrappedError struct {
	error
	frame Frame
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

func (we wrappedError) Unwrap() error {
	return we.error
}
