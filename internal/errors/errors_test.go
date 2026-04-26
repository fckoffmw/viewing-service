package errors

import (
	"testing"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		fn      func() *Error
		wantErr *Error
	}{
		{
			name:     "bad request",
			fn:       func() *Error { return BadRequest("invalid input") },
			wantErr:  &Error{Code: 400, Message: "invalid input"},
		},
		{
			name:     "unauthorized",
			fn:       func() *Error { return Unauthorized("unauthorized") },
			wantErr:  &Error{Code: 401, Message: "unauthorized"},
		},
		{
			name:     "not found",
			fn:       func() *Error { return NotFound("not found") },
			wantErr:  &Error{Code: 404, Message: "not found"},
		},
		{
			name:     "internal",
			fn:       func() *Error { return Internal("internal error") },
			wantErr:  &Error{Code: 500, Message: "internal error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err.Code != tt.wantErr.Code {
				t.Errorf("code = %d, want %d", err.Code, tt.wantErr.Code)
			}
			if err.Message != tt.wantErr.Message {
				t.Errorf("message = %q, want %q", err.Message, tt.wantErr.Message)
			}
		})
	}
}

func TestError(t *testing.T) {
	err := BadRequest("test error")
	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}