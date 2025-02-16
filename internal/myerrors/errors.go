package myerrors

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

var (
	UnknownError = ErrorMessage{Text: "unknown error type"}

	ErrLowBalance      = errors.New("not enough balance for the operation")
	ErrInvalidPassword = errors.New("invalid user password")
	ErrInvalidSender   = errors.New("non-empty sender value")
	ErrNullAmount      = errors.New("null amount value")
	ErrEmptyReceiver   = errors.New("empty receiver value")
)

type ErrorMessage struct {
	Text string `json:"errors"`
}

func New(v any) ErrorMessage {
	switch value := v.(type) {
	case string:
		return ErrorMessage{Text: value}
	case error:
		return ErrorMessage{Text: value.Error()}
	default:
		return UnknownError
	}
}

func Error(w http.ResponseWriter, r *http.Request, response any, status int) {
	render.Status(r, status)
	render.JSON(w, r, New(response))
}
