package apperror

type BusinessCode string

const (
	// General
	BE0000_UNKNOWN_ERROR           BusinessCode = "E0000"
	BE0001_INVALID_INPUT           BusinessCode = "E0001"
	BE0002_INPUT_VALIDATION_FAILED BusinessCode = "E0002"
	BE0003_UNAUTHORIZED            BusinessCode = "E0003"
	BE0004_FORBIDDEN               BusinessCode = "E0004"
	BE0005_RATE_LIMITED            BusinessCode = "E0005"

	// User/authentication
	BE1000_USER_ALREADY_EXISTS       BusinessCode = "E1000"
	BE1001_USER_NOT_FOUND            BusinessCode = "E1001"
	BE1002_RECOVERY_CODE_MISMATCH    BusinessCode = "E1002"
	BE1004_INACTIVE_USER             BusinessCode = "E1004"
	BE1006_CONFIRMATION_CODE_INVALID BusinessCode = "E1006"
	BE1007_CONFIRMATION_CODE_EXPIRED BusinessCode = "E1007"
	BE1008_ACCOUNT_ALREADY_CONFIRMED BusinessCode = "E1008"
	BE1009_ACCOUNT_NOT_CONFIRMED     BusinessCode = "E1009"
	BE1010_INVALID_CREDENTIALS       BusinessCode = "E1010"

	BE1011_REFRESH_TOKEN_REUSE_DETECTED BusinessCode = "E1011"
)
