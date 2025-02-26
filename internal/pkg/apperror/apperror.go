package apperror

type AppError interface {
	Error() string
	OriginalError() error
	Kind() Kind
	Origin() Origin
	OriginName() string
	BusinessCode() *BusinessCode
	ClientCode() *int
}

type appError struct {
	err          error
	message      string
	kind         Kind
	origin       Origin
	originName   string
	businessCode *BusinessCode
	clientCode   *int
}

func New(
	err error,
	msg string,
	kind Kind,
	origin Origin,
	originName string,
	bc *BusinessCode,
	clientCode *int,
) AppError {
	return &appError{
		err:          err,
		message:      msg,
		kind:         kind,
		origin:       origin,
		originName:   originName,
		businessCode: bc,
		clientCode:   clientCode,
	}
}

func (a *appError) Error() string {
	if a.message == "" {
		return a.err.Error()
	}

	return a.message
}

func (a *appError) OriginalError() error {
	return a.err
}

func (a *appError) Kind() Kind {
	return a.kind
}

func (a *appError) Origin() Origin {
	return a.origin
}

func (a *appError) OriginName() string {
	return a.originName
}

func (a *appError) BusinessCode() *BusinessCode {
	return a.businessCode
}

func (a *appError) ClientCode() *int {
	return a.clientCode
}
