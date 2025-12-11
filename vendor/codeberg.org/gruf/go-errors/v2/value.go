package errors

// WithValues is an optional interface that an error
// may implement to support key-value pair storage.
type WithValues interface {
	Value(key any) (value any)
}

// WithValue wraps err to store given key-value pair, accessible via Value() function.
func WithValue(err error, key any, value any) error {
	if err == nil {
		panic("nil error")
	}
	return &errorWithValue{err, key, value}
}

// Value searches for value stored under given key in error chain.
func Value(err error, key any) any {
	for errs := []error{err}; len(errs) > 0; {
		// Pop next error to check.
		err := errs[len(errs)-1]
		errs = errs[:len(errs)-1]
		if err == nil {
			continue
		}

		// Try extract value type.
		t, ok := err.(WithValues)
		if ok {
			if v := t.Value(key); v != nil {
				return v
			}
		}

		// Try unwrap any contained errors.
		if e := UnwrapV2(err); len(e) > 0 {
			errs = append(errs, e...)
		}
	}
	return nil
}

type errorWithValue struct {
	error
	k, v any
}

func (err *errorWithValue) Value(key any) any {
	if err.k == key {
		return err.v
	}
	if withValue, ok := err.error.(WithValues); ok {
		return withValue.Value(key)
	}
	return nil
}

func (err *errorWithValue) Unwrap() error {
	return err.error
}
