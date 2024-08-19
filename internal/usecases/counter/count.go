package counter

import "github.com/mathcale/go-api-boilerplate/internal/pkg/logger"

type CounterUseCase interface {
	Execute() (int, error)
}

type counterUseCase struct {
	logger  logger.Logger
	counter int
}

func NewCounterUseCase(l logger.Logger) CounterUseCase {
	return &counterUseCase{
		logger:  l,
		counter: 0,
	}
}

func (uc *counterUseCase) Execute() (int, error) {
	next := uc.counter + 1

	uc.logger.Info("[CounterUseCase] Incrementing counter", map[string]interface{}{
		"current_value": uc.counter,
		"next_value":    next,
	})

	uc.counter = next

	return next, nil
}
