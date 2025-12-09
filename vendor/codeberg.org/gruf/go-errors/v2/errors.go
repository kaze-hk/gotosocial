package errors

import (
	"fmt"
	"runtime"

	callers "codeberg.org/gruf/go-caller"
)

// New returns a new error created from message.
func New(msg string) error {
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		c.set(callers.Get(pcs[0]))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, StacktraceSize)
		n := runtime.Callers(2, pcs)
		t.set(pcs[:n])
	}
	return &errormsg{
		cfn: c,
		msg: msg,
		trc: t,
	}
}

// Newf returns a new error created from message format and args.
func Newf(msgf string, args ...interface{}) error {
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		c.set(callers.Get(pcs[0]))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, StacktraceSize)
		n := runtime.Callers(2, pcs)
		t.set(pcs[:n])
	}
	return &errormsg{
		cfn: c,
		msg: fmt.Sprintf(msgf, args...),
		trc: t,
	}
}

// NewAt returns a new error created, skipping 'skip'
// frames for trace / caller information, from message.
func NewAt(skip int, msg string) error {
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(skip+1, pcs)
		c.set(callers.Get(pcs[0]))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, StacktraceSize)
		n := runtime.Callers(skip+1, pcs)
		t.set(pcs[:n])
	}
	return &errormsg{
		cfn: c,
		msg: msg,
		trc: t,
	}
}

// Wrap will wrap supplied error within a new error created from message.
func Wrap(err error, msg string) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		c.set(callers.Get(pcs[0]))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, StacktraceSize)
		n := runtime.Callers(2, pcs)
		t.set(pcs[:n])
	}
	return &errorwrap{
		cfn: c,
		msg: msg,
		err: err,
		trc: t,
	}
}

// Wrapf will wrap supplied error within a new error created from message format and args.
func Wrapf(err error, msgf string, args ...interface{}) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(2, pcs)
		c.set(callers.Get(pcs[0]))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, StacktraceSize)
		n := runtime.Callers(2, pcs)
		t.set(pcs[:n])
	}
	return &errorwrap{
		cfn: c,
		msg: fmt.Sprintf(msgf, args...),
		err: err,
		trc: t,
	}
}

// WrapAt wraps error within new error created from message,
// skipping 'skip' frames for trace / caller information.
func WrapAt(skip int, err error, msg string) error {
	if err == nil {
		panic("cannot wrap nil error")
	}
	var c caller
	var t trace
	if IncludesCaller {
		pcs := make([]uintptr, 1)
		_ = runtime.Callers(skip+1, pcs)
		c.set(callers.Get(pcs[0]))
	}
	if IncludesStacktrace {
		pcs := make([]uintptr, StacktraceSize)
		n := runtime.Callers(skip+1, pcs)
		t.set(pcs[:n])
	}
	return &errorwrap{
		cfn: c,
		msg: msg,
		err: err,
		trc: t,
	}
}

// Stacktrace fetches first stored stacktrace of callers from error chain.
func Stacktrace(err error) Callers {
	if !IncludesStacktrace {
		// compile-time check
		return nil
	}

	// Check for usable type.
	switch t := err.(type) {
	case nil:
		return nil
	case *errormsg:
		return t.trc.value()
	case *errorwrap:
		return t.trc.value()
	case interface{ As(any) bool }:
		if e := new(errormsg); t.As(&e) {
			return e.trc.value()
		}
		if e := new(errorwrap); t.As(&e) {
			return e.trc.value()
		}
	}

	// Above is fast-path without
	// needing to allocate a slice
	// or enter a loop.
	var errs []error

	// Try unwrap errors.
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		errs = append(errs, u.Unwrap())
	case interface{ Unwrap() []error }:
		errs = append(errs, u.Unwrap()...)
	}

	for len(errs) > 0 {
		// Pop next error to check.
		err := errs[len(errs)-1]
		errs = errs[:len(errs)-1]

		// Handle depending on type.
		switch t := err.(type) {
		case nil:
			return nil
		case *errormsg:
			return t.trc.value()
		case *errorwrap:
			return t.trc.value()
		case interface{ As(any) bool }:
			if e := new(errormsg); t.As(&e) {
				return e.trc.value()
			}
			if e := new(errorwrap); t.As(&e) {
				return e.trc.value()
			}
		case interface{ Unwrap() error }:
			errs = append(errs, t.Unwrap())
		case interface{ Unwrap() []error }:
			errs = append(errs, t.Unwrap()...)
		}
	}

	return nil
}

type errormsg struct {
	cfn caller
	msg string
	trc trace
}

func (err *errormsg) Error() string {
	if IncludesCaller {
		fn := err.cfn.value()
		return fn + " " + err.msg
	} else {
		return err.msg
	}
}

func (err *errormsg) Is(other error) bool {
	oerr, ok := other.(*errormsg)
	return ok && oerr.msg == err.msg
}

type errorwrap struct {
	cfn caller
	msg string
	err error // wrapped
	trc trace
}

func (err *errorwrap) Error() string {
	if IncludesCaller {
		fn := err.cfn.value()
		return fn + " " + err.msg + ": " + err.err.Error()
	} else {
		return err.msg + ": " + err.err.Error()
	}
}

func (err *errorwrap) Is(other error) bool {
	oerr, ok := other.(*errorwrap)
	return ok && oerr.msg == err.msg
}

func (err *errorwrap) Unwrap() error {
	return err.err
}
