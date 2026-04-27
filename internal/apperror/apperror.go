package apperror

import "errors"

type Detail struct {
	Campo    string `json:"campo"`
	Mensagem string `json:"mensagem"`
}

type Error struct {
	Status   int      `json:"-"`
	Codigo   string   `json:"codigo"`
	Mensagem string   `json:"mensagem"`
	Detalhes []Detail `json:"detalhes,omitempty"`
	err      error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	return e.Mensagem
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.err
}

func New(status int, codigo string, mensagem string) *Error {
	return &Error{Status: status, Codigo: codigo, Mensagem: mensagem}
}

func WithDetails(err *Error, detalhes ...Detail) *Error {
	err.Detalhes = detalhes
	return err
}

func Wrap(status int, codigo string, mensagem string, err error) *Error {
	return &Error{Status: status, Codigo: codigo, Mensagem: mensagem, err: err}
}

func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}
