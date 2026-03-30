// pkg/i18n/middleware.go
package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/835-droid/ms-ai-backend/pkg/errors"
)

// LanguageContextKey is the key used to store language in gin context
const LanguageContextKey = "language"

// LanguageMiddleware creates middleware to detect and set user language
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		translator := GetTranslator()

		// Priority order for language detection:
		// 1. Query parameter: ?lang=en
		// 2. Accept-Language header
		// 3. Default language

		var lang Language

		// Check query parameter first
		if queryLang := c.Query("lang"); queryLang != "" {
			if translator.IsLanguageSupported(Language(queryLang)) {
				lang = Language(queryLang)
			}
		}

		// If not set, check Accept-Language header
		if lang == "" {
			acceptLang := c.GetHeader("Accept-Language")
			if acceptLang != "" {
				// Parse Accept-Language header (e.g., "en-US,en;q=0.9,ar;q=0.8")
				lang = parseAcceptLanguage(acceptLang, translator)
			}
		}

		// If still not set, use default language
		if lang == "" {
			lang = translator.GetDefaultLanguage()
		}

		// Store language in context
		c.Set(LanguageContextKey, lang)

		c.Next()
	}
}

// parseAcceptLanguage parses Accept-Language header and returns best matching supported language
func parseAcceptLanguage(acceptLang string, translator *Translator) Language {
	// Split by comma and process each language tag
	langs := strings.Split(acceptLang, ",")

	for _, langTag := range langs {
		// Remove quality value (e.g., ";q=0.9")
		lang := strings.Split(strings.TrimSpace(langTag), ";")[0]

		// Extract primary language (e.g., "en" from "en-US")
		primaryLang := strings.Split(lang, "-")[0]

		// Check if primary language is supported
		if translator.IsLanguageSupported(Language(primaryLang)) {
			return Language(primaryLang)
		}

		// Check if full language tag is supported
		if translator.IsLanguageSupported(Language(lang)) {
			return Language(lang)
		}
	}

	return "" // No supported language found
}

// GetLanguageFromContext retrieves the language from gin context
func GetLanguageFromContext(c *gin.Context) Language {
	if lang, exists := c.Get(LanguageContextKey); exists {
		if language, ok := lang.(Language); ok {
			return language
		}
	}
	return GetTranslator().GetDefaultLanguage()
}

// TContext is a shorthand for translating using language from context
func TContext(c *gin.Context, key MessageKey, args ...interface{}) string {
	lang := GetLanguageFromContext(c)
	return GetTranslator().Translate(lang, key, args...)
}

// TEContext is a shorthand for translating error using language from context
func TEContext(c *gin.Context, err *errors.DomainError) *errors.DomainError {
	lang := GetLanguageFromContext(c)
	return GetTranslator().TranslateError(lang, err)
}
