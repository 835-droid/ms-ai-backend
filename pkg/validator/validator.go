package validator

import (
	"errors"
	"regexp"

	// استيراد الحزمة الجديدة للتحقق من الـ Structs
	"github.com/go-playground/validator/v10"
)

// The global validator instance for struct validation
var validate *validator.Validate

func init() {
	// Initialize the validator instance once
	validate = validator.New()
}

// 🚀 الدالة المفقودة التي تسببت في خطأ "undefined"
// Validate takes any struct and performs validation based on its tags (e.g., validate:"required,min=5").
// It returns an error if validation fails.
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		// يمكن هنا تعديل رسائل الخطأ لتكون أكثر وضوحاً
		return err
	}
	return nil
}

// الدوال الحالية الخاصة بك (ValidateUsername, ValidatePassword, ValidateInviteCode)

var (
	usernameRE  = regexp.MustCompile(`^[A-Za-z_\-][A-Za-z0-9_\-]{2,29}$`)
	uppercaseRE = regexp.MustCompile(`[A-Z]`)
	lowercaseRE = regexp.MustCompile(`[a-z]`)
	digitRE     = regexp.MustCompile(`[0-9]`)
	specialRE   = regexp.MustCompile(`[!@#\$%\^&\*()_+\-=\[\]{};':"\\|,.<>\/?]`)
)

func ValidateUsername(u string) error {
	if u == "" {
		return errors.New("username is required")
	}
	if !usernameRE.MatchString(u) {
		return errors.New("username must be 3-30 characters, start with a letter/underscore/dash and contain only alphanumeric, underscore or dash")
	}
	return nil
}

func ValidatePassword(p string) error {
	if len(p) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !uppercaseRE.MatchString(p) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !lowercaseRE.MatchString(p) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !digitRE.MatchString(p) {
		return errors.New("password must contain at least one digit")
	}
	if !specialRE.MatchString(p) {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

func ValidateInviteCode(code string) error {
	if code == "" {
		return errors.New("invite code is required")
	}
	// basic check: only allowed charset
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return errors.New("invalid invite code format")
	}
	return nil
}
