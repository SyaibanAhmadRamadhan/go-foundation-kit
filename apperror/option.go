package apperror

type Option func(*Error)

// WithPublicMessage mengatur pesan yang akan dikirim ke user.
func WithPublicMessage(msg string) Option {
	return func(e *Error) {
		e.PublicMessage = msg
	}
}

func WithStack() Option {
	return func(e *Error) {
		e.Stack = Stack()
	}
}

func WithCause(err error) Option {
	return func(e *Error) { e.Cause = err }
}
