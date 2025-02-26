package apperror

type Kind string

const (
	ParseErrorKind Kind = "parse_error"
	ValidationKind Kind = "validation_error"
	RestClientKind Kind = "rest_client_error"
	DatabaseKind   Kind = "database_error"
	BusinessKind   Kind = "business_error"
)
