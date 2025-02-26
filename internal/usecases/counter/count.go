package counter

import (
	"errors"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type CounterUseCase interface {
	Execute() (int, error)
}

type counterUseCase struct {
	name         string
	logger       logger.Logger
	counterValue int
}

func NewCounterUseCase(l logger.Logger) CounterUseCase {
	return &counterUseCase{
		name:         "counter",
		logger:       l,
		counterValue: 0,
	}
}

func (uc *counterUseCase) Execute() (int, error) {
	next := uc.counterValue + 1

	uc.logger.Debug("Incrementing counter", map[string]interface{}{
		"use_case":      uc.name,
		"current_value": uc.counterValue,
		"next_value":    next,
	})

	if next > 3 {
		msg := "maximum counter value reached"
		businessCode := apperror.BUSINESS_E100

		return uc.counterValue, apperror.New(
			errors.New(msg), msg,
			apperror.BusinessKind, apperror.UseCaseOrigin, uc.name, &businessCode, nil,
		)
	}

	uc.counterValue = next

	return next, nil
}
