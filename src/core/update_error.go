package core

type UpdateError struct {
	cause error
}

func NewUpdateError(cause error) UpdateError {
	return UpdateError{
		cause: cause,
	}
}

func (e UpdateError) Error() string {
	return e.cause.Error()
}

func (e UpdateError) Unwrap() error {
	return e.cause
}

var Err_updateError = &UpdateError{}
