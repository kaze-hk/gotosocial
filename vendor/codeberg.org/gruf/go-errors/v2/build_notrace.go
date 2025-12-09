//go:build !errtrace
// +build !errtrace

package errors

// IncludesStacktrace is a compile-time flag used to indicate
// whether to include stacktraces on error wrap / creation.
const IncludesStacktrace = false

// StacktraceSize allows adjusting
// the stacktrace size at runtime.
var StacktraceSize uint = 10

type trace struct{}

// set will set the actual trace value
// only when correct build flag is set.
func (trace) set([]uintptr) {}

// value returns the actual trace value
// only when correct build flag is set.
func (trace) value() Callers { return nil }
