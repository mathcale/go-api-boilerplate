package apperror

type Origin string

const (
	WebHandlerOrigin     Origin = "web_handler"
	UseCaseOrigin        Origin = "use_case"
	UseCaseGatewayOrigin Origin = "gateway"
	RepositoryOrigin     Origin = "repository"
	MiddlewareOrigin     Origin = "middleware"
	PackageOrigin        Origin = "pkg"
)
