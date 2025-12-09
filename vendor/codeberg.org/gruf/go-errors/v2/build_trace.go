//go:build errtrace
// +build errtrace

package errors

import (
	"runtime"
)

// IncludesStacktrace is a compile-time flag used to indicate
// whether to include stacktraces on error wrap / creation.
const IncludesStacktrace = true

// StacktraceSize allows adjusting
// the stacktrace size at runtime.
var StacktraceSize uint = 10

type trace []uintptr

// set will set the actual trace value
// only when correct build flag is set.
func (t *trace) set(v []uintptr) {
	*t = trace(v)
}

// value returns the actual trace value
// only when correct build flag is set.
func (t trace) value() Callers {
	iter := runtime.CallersFrames([]uintptr(t))
	return Callers(gatherFrames(iter, len(t)))
}

// gatherFrames collates runtime frames from a frame iterator.
func gatherFrames(iter *runtime.Frames, n int) Callers {
	if iter == nil {
		return nil
	}
	frames := make([]runtime.Frame, 0, n)
	for {
		f, ok := iter.Next()
		if !ok {
			break
		}
		frames = append(frames, f)
	}
	return frames
}
