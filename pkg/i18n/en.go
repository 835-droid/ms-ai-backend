// pkg/i18n/en.go
package i18n

// getEnglishMessages returns English message translations
func getEnglishMessages() map[MessageKey]string {
	return map[MessageKey]string{
		// Authentication messages
		MsgAuthLoginSuccess:       "Login successful",
		MsgAuthLoginFailed:        "Login failed",
		MsgAuthLogoutSuccess:      "Logout successful",
		MsgAuthSignupSuccess:      "Account created successfully",
		MsgAuthUserExists:         "User already exists",
		MsgAuthInvalidCredentials: "Invalid email or password",
		MsgAuthTokenExpired:       "Authentication token has expired",
		MsgAuthTokenInvalid:       "Invalid authentication token",
		MsgAuthAccessDenied:       "Access denied",
		MsgAuthInsufficientRole:   "Insufficient permissions for this action",

		// Validation messages
		MsgValidationRequired:      "This field is required",
		MsgValidationMinLength:     "Minimum length is %d characters",
		MsgValidationMaxLength:     "Maximum length is %d characters",
		MsgValidationInvalidFormat: "Invalid format",
		MsgValidationInvalidEmail:  "Invalid email address",
		MsgValidationPasswordWeak:  "Password is too weak. Must contain at least 8 characters, one uppercase letter, one lowercase letter, one number, and one special character",

		// User messages
		MsgUserNotFound:       "User not found",
		MsgUserProfileUpdated: "Profile updated successfully",
		MsgUserDeleted:        "User account deleted successfully",

		// Manga messages
		MsgMangaCreated:         "Manga created successfully",
		MsgMangaUpdated:         "Manga updated successfully",
		MsgMangaDeleted:         "Manga deleted successfully",
		MsgMangaNotFound:        "Manga not found",
		MsgMangaChapterAdded:    "Chapter added successfully",
		MsgMangaChapterNotFound: "Chapter not found",

		// System messages
		MsgSystemInternalError:      "Internal server error",
		MsgSystemServiceUnavailable: "Service temporarily unavailable",
		MsgSystemRateLimitExceeded:  "Too many requests. Please try again later",
		MsgSystemMaintenance:        "System is under maintenance",

		// Invite codes
		MsgInviteCodeUsed:      "Invite code has already been used",
		MsgInviteCodeExpired:   "Invite code has expired",
		MsgInviteCodeInvalid:   "Invalid invite code",
		MsgInviteCodeGenerated: "Invite code generated successfully",
	}
}
