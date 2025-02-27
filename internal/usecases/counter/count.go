package counter

import (
	"errors"
	"fmt"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type CounterUseCase interface {
	Execute(limit int) (int, error)
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

func (uc *counterUseCase) Execute(limit int) (int, error) {
	next := uc.counterValue + 1

	uc.logger.Debug("Incrementing counter", map[string]interface{}{
		"use_case":      uc.name,
		"current_value": uc.counterValue,
		"next_value":    next,
	})

	if next > limit {
		msg := fmt.Sprintf("maximum counter value of [%d] reached", limit)
		businessCode := apperror.BUSINESS_E100

		return uc.counterValue, apperror.New(
			errors.New(msg), msg,
			apperror.BusinessKind, apperror.UseCaseOrigin, uc.name, &businessCode, nil,
		)
	}

	uc.counterValue = next

	return next, nil
}
