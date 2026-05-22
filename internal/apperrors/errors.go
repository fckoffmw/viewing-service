package apperrors

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (e *Error) Error() string {
	return e.Message
}

func BadRequest(msg string) *Error {
	return &Error{Code: 400, Message: msg}
}

func Unauthorized(msg string) *Error {
	return &Error{Code: 401, Message: msg}
}

func NotFound(msg string) *Error {
	return &Error{Code: 404, Message: msg}
}

func Internal(msg string) *Error {
	return &Error{Code: 500, Message: msg}
}