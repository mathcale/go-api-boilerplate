package apperror

type Kind string

const (
	ParseKind      Kind = "parse_error"
	ValidationKind Kind = "validation_error"
	RestClientKind Kind = "rest_client_error"
	DatabaseKind   Kind = "database_error"
	BusinessKind   Kind = "business_error"
	DependencyKind Kind = "dependency_error"
	ConflictKind   Kind = "conflict_error"
)
