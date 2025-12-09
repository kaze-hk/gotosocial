package errors

import (
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// Callers is a wrapper around []runtime.Frame
// to add useful serialization functions.
type Callers []runtime.Frame

// MarshalJSON implements json.Marshaler
// to provide an easy, simple default.
func (c Callers) MarshalJSON() ([]byte, error) {
	type jsonFrame struct {
		Func string `json:"func"`
		File string `json:"file"`
		Line int    `json:"line"`
	}

	// Allocate expected size jsonFrame slice.
	jsonFrames := make([]jsonFrame, len(c))
	if len(jsonFrames) != len(c) {
		panic("bound check elimination")
	}

	// Convert each to jsonFrame object.
	for i := 0; i < len(c); i++ {
		frame := c[i]
		jsonFrames[i] = jsonFrame{
			Func: funcName(frame.Func),
			File: frame.File,
			Line: frame.Line,
		}
	}

	// Marshal converted frames.
	return json.Marshal(jsonFrames)
}

// String will return a simple string
// representation of receiving Callers slice.
func (c Callers) String() string {

	// Guess-timate to reduce allocs.
	buf := make([]byte, 0, 64*len(c))
	for i := 0; i < len(c); i++ {
		frame := c[i]

		// Append formatted caller info.
		fn := funcName(frame.Func)
		buf = append(buf, fn+"()\n\t"+frame.File+":"...)
		buf = strconv.AppendInt(buf, int64(frame.Line), 10)
		buf = append(buf, '\n')
	}

	return unsafe.String(&buf[0], len(buf))
}

// funcName formats a function name to a quickly-readable string.
func funcName(fn *runtime.Func) string {
	if fn == nil {
		return ""
	}

	// Get func name
	// for formatting.
	name := fn.Name()

	// Drop all but the package name and function name, no mod path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	const params = `[...]`

	// Drop any generic type parameter markers
	if idx := strings.Index(name, params); idx >= 0 {
		name = name[:idx] + name[idx+len(params):]
	}

	return name
}
