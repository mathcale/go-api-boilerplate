package apperror

type BusinessCode string

// Tip: map these to your business errors so it's easier for your client to report what happened
const (
	BUSINESS_E100 BusinessCode = "E100"
	BUSINESS_E101 BusinessCode = "E101"
	BUSINESS_E102 BusinessCode = "E102"
	BUSINESS_E103 BusinessCode = "E103"
	BUSINESS_E104 BusinessCode = "E104"
	BUSINESS_E105 BusinessCode = "E105"
)
