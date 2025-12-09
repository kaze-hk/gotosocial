package errors

// WithValue wraps err to store given key-value pair, accessible via Value() function.
func WithValue(err error, key any, value any) error {
	if err == nil {
		panic("nil error")
	}
	var kvs []kv
	if e := asErrWithValues(err); e != nil {
		kvs = e.kvs
	}
	return &errWithValues{
		err: err,
		kvs: append(kvs, kv{key, value}),
	}
}

// Value searches for value stored under given key in error chain.
func Value(err error, key any) any {
	if e := asErrWithValues(err); e != nil {
		for _, kv := range e.kvs {
			if kv.k == key {
				return kv.v
			}
		}
	}
	return nil
}

// simple key-value type.
type kv struct{ k, v any }

// errWithValues wraps an error
// to provide key-value storage.
type errWithValues struct {
	err error
	kvs []kv
}

func (e *errWithValues) Error() string {
	return e.err.Error()
}

func (e *errWithValues) Unwrap() error {
	return e.err
}

// asErrWithValues is a manually monomorphized form
// of AsV2[*errWithValues] to improve performance.
func asErrWithValues(err error) *errWithValues {
	if err == nil {
		return nil
	}

	// Check for direct type match.
	t, ok := err.(*errWithValues)
	if ok {
		return t
	}

	// Look for .As() support.
	as, ok := err.(interface {
		As(target any) bool
	})

	// Try to call .As().
	if ok && as.As(&t) {
		return t
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
		if err == nil {
			continue
		}

		// Check for direct type match.
		t, ok := err.(*errWithValues)
		if ok {
			return t
		}

		// Look for .As() support.
		as, ok := err.(interface {
			As(target any) bool
		})

		// Try to call .As().
		if ok && as.As(&t) {
			return t
		}

		// Try unwrap errors.
		switch u := err.(type) {
		case interface{ Unwrap() error }:
			errs = append(errs, u.Unwrap())
		case interface{ Unwrap() []error }:
			errs = append(errs, u.Unwrap()...)
		}
	}

	return nil
}
