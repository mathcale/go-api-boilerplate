package dto

import "github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"

type ErrorOutput struct {
	Message string                 `json:"message"`
	Code    *apperror.BusinessCode `json:"code,omitempty"`
	Cause   string                 `json:"cause,omitempty"`
}
