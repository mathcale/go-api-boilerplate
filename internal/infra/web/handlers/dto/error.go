package dto

import "github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"

type ErrorOutput struct {
	Code    *apperror.BusinessCode `json:"code"`
	Message string                 `json:"message"`
}
