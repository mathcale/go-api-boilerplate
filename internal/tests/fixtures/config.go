package fixtures

import "strings"

var (
	ConfigAccessSecret   = strings.Repeat("a1B2c3D4", 8)
	ConfigRefreshSecret  = strings.Repeat("e5F6g7H8", 8)
	ConfigSMTPCredential = strings.Repeat("s9K3", 3)
)

const (
	ConfigPlaceholderAccess  = "change-me-access-secret"
	ConfigPlaceholderRefresh = "change-me-refresh-secret"
	ConfigShortValue         = "too-short-secret"
)
