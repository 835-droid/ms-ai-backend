// pkg/i18n/i18n.go
package i18n

import (
	"fmt"
	"sync"

	"github.com/835-droid/ms-ai-backend/pkg/errors"
)

// Language represents a supported language
type Language string

const (
	// Supported languages
	LanguageEnglish Language = "en"
	LanguageArabic  Language = "ar"
	LanguageFrench  Language = "fr"
	LanguageSpanish Language = "es"
	LanguageGerman  Language = "de"
)

// MessageKey represents a message identifier
type MessageKey string

const (
	// Authentication messages
	MsgAuthLoginSuccess       MessageKey = "auth.login.success"
	MsgAuthLoginFailed        MessageKey = "auth.login.failed"
	MsgAuthLogoutSuccess      MessageKey = "auth.logout.success"
	MsgAuthSignupSuccess      MessageKey = "auth.signup.success"
	MsgAuthUserExists         MessageKey = "auth.user.exists"
	MsgAuthInvalidCredentials MessageKey = "auth.invalid.credentials"
	MsgAuthTokenExpired       MessageKey = "auth.token.expired"
	MsgAuthTokenInvalid       MessageKey = "auth.token.invalid"
	MsgAuthAccessDenied       MessageKey = "auth.access.denied"
	MsgAuthInsufficientRole   MessageKey = "auth.insufficient.role"

	// Validation messages
	MsgValidationRequired      MessageKey = "validation.required"
	MsgValidationMinLength     MessageKey = "validation.min_length"
	MsgValidationMaxLength     MessageKey = "validation.max_length"
	MsgValidationInvalidFormat MessageKey = "validation.invalid_format"
	MsgValidationInvalidEmail  MessageKey = "validation.invalid_email"
	MsgValidationPasswordWeak  MessageKey = "validation.password_weak"

	// User messages
	MsgUserNotFound       MessageKey = "user.not_found"
	MsgUserProfileUpdated MessageKey = "user.profile.updated"
	MsgUserDeleted        MessageKey = "user.deleted"

	// Manga messages
	MsgMangaCreated         MessageKey = "manga.created"
	MsgMangaUpdated         MessageKey = "manga.updated"
	MsgMangaDeleted         MessageKey = "manga.deleted"
	MsgMangaNotFound        MessageKey = "manga.not_found"
	MsgMangaChapterAdded    MessageKey = "manga.chapter.added"
	MsgMangaChapterNotFound MessageKey = "manga.chapter.not_found"

	// System messages
	MsgSystemInternalError      MessageKey = "system.internal_error"
	MsgSystemServiceUnavailable MessageKey = "system.service_unavailable"
	MsgSystemRateLimitExceeded  MessageKey = "system.rate_limit_exceeded"
	MsgSystemMaintenance        MessageKey = "system.maintenance"

	// Invite codes
	MsgInviteCodeUsed      MessageKey = "invite.code.used"
	MsgInviteCodeExpired   MessageKey = "invite.code.expired"
	MsgInviteCodeInvalid   MessageKey = "invite.code.invalid"
	MsgInviteCodeGenerated MessageKey = "invite.code.generated"
)

// Translator provides internationalization functionality
type Translator struct {
	mu          sync.RWMutex
	messages    map[Language]map[MessageKey]string
	defaultLang Language
}

// NewTranslator creates a new translator instance
func NewTranslator(defaultLang Language) *Translator {
	t := &Translator{
		messages:    make(map[Language]map[MessageKey]string),
		defaultLang: defaultLang,
	}

	// Load all language files
	t.loadLanguages()

	return t
}

// loadLanguages loads all language message files
func (t *Translator) loadLanguages() {
	// Load English messages
	t.messages[LanguageEnglish] = getEnglishMessages()

	// Load Arabic messages
	t.messages[LanguageArabic] = getArabicMessages()

	// Load French messages
	t.messages[LanguageFrench] = getFrenchMessages()

	// Load Spanish messages
	t.messages[LanguageSpanish] = getSpanishMessages()

	// Load German messages
	t.messages[LanguageGerman] = getGermanMessages()
}

// Translate translates a message key to the specified language
func (t *Translator) Translate(lang Language, key MessageKey, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try requested language first
	if langMessages, exists := t.messages[lang]; exists {
		if message, found := langMessages[key]; found {
			if len(args) > 0 {
				return fmt.Sprintf(message, args...)
			}
			return message
		}
	}

	// Fallback to default language
	if defaultMessages, exists := t.messages[t.defaultLang]; exists {
		if message, found := defaultMessages[key]; found {
			if len(args) > 0 {
				return fmt.Sprintf(message, args...)
			}
			return message
		}
	}

	// Final fallback - return key as string
	return string(key)
}

// TranslateError translates a domain error to the specified language
func (t *Translator) TranslateError(lang Language, err *errors.DomainError) *errors.DomainError {
	if err == nil {
		return nil
	}

	// Map error codes to message keys
	var messageKey MessageKey
	switch err.Code {
	case errors.ErrCodeAuthenticationFailed:
		messageKey = MsgAuthLoginFailed
	case errors.ErrCodeInvalidCredentials:
		messageKey = MsgAuthInvalidCredentials
	case errors.ErrCodeTokenExpired:
		messageKey = MsgAuthTokenExpired
	case errors.ErrCodeTokenInvalid:
		messageKey = MsgAuthTokenInvalid
	case errors.ErrCodeAccessDenied:
		messageKey = MsgAuthAccessDenied
	case errors.ErrCodeInsufficientRole:
		messageKey = MsgAuthInsufficientRole
	case errors.ErrCodeUserNotFound:
		messageKey = MsgUserNotFound
	case errors.ErrCodeUserExists:
		messageKey = MsgAuthUserExists
	case errors.ErrCodeNotFound:
		messageKey = MsgMangaNotFound
	case errors.ErrCodeInternalServer:
		messageKey = MsgSystemInternalError
	default:
		// Return original error if no translation found
		return err
	}

	translatedMessage := t.Translate(lang, messageKey)

	// Create new error with translated message
	return &errors.DomainError{
		Code:       err.Code,
		Message:    translatedMessage,
		Details:    err.Details,
		HTTPStatus: err.HTTPStatus,
		Cause:      err.Cause,
	}
}

// GetSupportedLanguages returns list of supported languages
func (t *Translator) GetSupportedLanguages() []Language {
	t.mu.RLock()
	defer t.mu.RUnlock()

	languages := make([]Language, 0, len(t.messages))
	for lang := range t.messages {
		languages = append(languages, lang)
	}
	return languages
}

// IsLanguageSupported checks if a language is supported
func (t *Translator) IsLanguageSupported(lang Language) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	_, exists := t.messages[lang]
	return exists
}

// GetDefaultLanguage returns the default language
func (t *Translator) GetDefaultLanguage() Language {
	return t.defaultLang
}

// SetDefaultLanguage sets the default language
func (t *Translator) SetDefaultLanguage(lang Language) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.IsLanguageSupported(lang) {
		t.defaultLang = lang
	}
}

// AddMessage adds a custom message for a language and key
func (t *Translator) AddMessage(lang Language, key MessageKey, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.messages[lang] == nil {
		t.messages[lang] = make(map[MessageKey]string)
	}
	t.messages[lang][key] = message
}

// Global translator instance
var globalTranslator *Translator
var translatorOnce sync.Once

// GetTranslator returns the global translator instance
func GetTranslator() *Translator {
	translatorOnce.Do(func() {
		globalTranslator = NewTranslator(LanguageEnglish)
	})
	return globalTranslator
}

// T is a shorthand for GetTranslator().Translate()
func T(lang Language, key MessageKey, args ...interface{}) string {
	return GetTranslator().Translate(lang, key, args...)
}

// TE is a shorthand for GetTranslator().TranslateError()
func TE(lang Language, err *errors.DomainError) *errors.DomainError {
	return GetTranslator().TranslateError(lang, err)
}
