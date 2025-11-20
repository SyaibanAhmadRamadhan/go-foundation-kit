package apperror

import "runtime/debug"

type Option func(*Error)

// WithPublicMessage mengatur pesan yang akan dikirim ke user.
func WithPublicMessage(msg string) Option {
	return func(e *Error) {
		e.PublicMessage = msg
	}
}

func WithStack() Option {
	return func(e *Error) {
		e.Stack = string(debug.Stack())
	}
}

func EnableStack(enable bool) Option {
	return func(e *Error) {
		if enable {
			e.Stack = string(debug.Stack())
		} else {
			e.Stack = ""
		}
	}
}
