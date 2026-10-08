package apierror

import (
	"fmt"

	"github.com/mathcale/go-api-boilerplate/internal/core/apperror"
)

type (
	ApiError interface {
		Error() string
		Status() int
		Message() string
		BusinessCode() *apperror.BusinessCode
	}

	apiError struct {
		err          error
		status       int
		message      string
		businessCode *apperror.BusinessCode
	}
)

func New(err error, status int, msg string, bc *apperror.BusinessCode) ApiError {
	return &apiError{
		err:          err,
		status:       status,
		message:      msg,
		businessCode: bc,
	}
}

func (a *apiError) Error() string {
	businessCodeStr := ""

	if a.businessCode != nil {
		businessCodeStr = string(*a.businessCode)
	}

	return fmt.Sprintf(
		"message:[%s] | status:[%d] | business_code:[%s] | error:[%s]",
		a.message, a.status, businessCodeStr, a.err,
	)
}

func (a *apiError) Status() int {
	return a.status
}

func (a *apiError) Message() string {
	return a.message
}

func (a *apiError) BusinessCode() *apperror.BusinessCode {
	return a.businessCode
}
