// pkg/i18n/de.go
package i18n

// getGermanMessages returns German message translations
func getGermanMessages() map[MessageKey]string {
	return map[MessageKey]string{
		// Authentication messages
		MsgAuthLoginSuccess:       "Anmeldung erfolgreich",
		MsgAuthLoginFailed:        "Anmeldung fehlgeschlagen",
		MsgAuthLogoutSuccess:      "Abmeldung erfolgreich",
		MsgAuthSignupSuccess:      "Konto erfolgreich erstellt",
		MsgAuthUserExists:         "Benutzer existiert bereits",
		MsgAuthInvalidCredentials: "Ungültige E-Mail oder Passwort",
		MsgAuthTokenExpired:       "Authentifizierungstoken ist abgelaufen",
		MsgAuthTokenInvalid:       "Ungültiger Authentifizierungstoken",
		MsgAuthAccessDenied:       "Zugriff verweigert",
		MsgAuthInsufficientRole:   "Unzureichende Berechtigungen für diese Aktion",

		// Validation messages
		MsgValidationRequired:      "Dieses Feld ist erforderlich",
		MsgValidationMinLength:     "Mindestlänge beträgt %d Zeichen",
		MsgValidationMaxLength:     "Maximallänge beträgt %d Zeichen",
		MsgValidationInvalidFormat: "Ungültiges Format",
		MsgValidationInvalidEmail:  "Ungültige E-Mail-Adresse",
		MsgValidationPasswordWeak:  "Das Passwort ist zu schwach. Muss mindestens 8 Zeichen, einen Großbuchstaben, einen Kleinbuchstaben, eine Zahl und ein Sonderzeichen enthalten",

		// User messages
		MsgUserNotFound:       "Benutzer nicht gefunden",
		MsgUserProfileUpdated: "Profil erfolgreich aktualisiert",
		MsgUserDeleted:        "Benutzerkonto erfolgreich gelöscht",

		// Manga messages
		MsgMangaCreated:         "Manga erfolgreich erstellt",
		MsgMangaUpdated:         "Manga erfolgreich aktualisiert",
		MsgMangaDeleted:         "Manga erfolgreich gelöscht",
		MsgMangaNotFound:        "Manga nicht gefunden",
		MsgMangaChapterAdded:    "Kapitel erfolgreich hinzugefügt",
		MsgMangaChapterNotFound: "Kapitel nicht gefunden",

		// System messages
		MsgSystemInternalError:      "Interner Serverfehler",
		MsgSystemServiceUnavailable: "Dienst vorübergehend nicht verfügbar",
		MsgSystemRateLimitExceeded:  "Zu viele Anfragen. Bitte versuchen Sie es später erneut",
		MsgSystemMaintenance:        "System wird gewartet",

		// Invite codes
		MsgInviteCodeUsed:      "Einladungscode wurde bereits verwendet",
		MsgInviteCodeExpired:   "Einladungscode ist abgelaufen",
		MsgInviteCodeInvalid:   "Ungültiger Einladungscode",
		MsgInviteCodeGenerated: "Einladungscode erfolgreich generiert",
	}
}
