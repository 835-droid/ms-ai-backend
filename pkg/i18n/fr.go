// pkg/i18n/fr.go
package i18n

// getFrenchMessages returns French message translations
func getFrenchMessages() map[MessageKey]string {
	return map[MessageKey]string{
		// Authentication messages
		MsgAuthLoginSuccess:       "Connexion réussie",
		MsgAuthLoginFailed:        "Échec de la connexion",
		MsgAuthLogoutSuccess:      "Déconnexion réussie",
		MsgAuthSignupSuccess:      "Compte créé avec succès",
		MsgAuthUserExists:         "L'utilisateur existe déjà",
		MsgAuthInvalidCredentials: "Email ou mot de passe invalide",
		MsgAuthTokenExpired:       "Le jeton d'authentification a expiré",
		MsgAuthTokenInvalid:       "Jeton d'authentification invalide",
		MsgAuthAccessDenied:       "Accès refusé",
		MsgAuthInsufficientRole:   "Permissions insuffisantes pour cette action",

		// Validation messages
		MsgValidationRequired:      "Ce champ est requis",
		MsgValidationMinLength:     "La longueur minimale est de %d caractères",
		MsgValidationMaxLength:     "La longueur maximale est de %d caractères",
		MsgValidationInvalidFormat: "Format invalide",
		MsgValidationInvalidEmail:  "Adresse email invalide",
		MsgValidationPasswordWeak:  "Le mot de passe est trop faible. Doit contenir au moins 8 caractères, une lettre majuscule, une lettre minuscule, un chiffre et un caractère spécial",

		// User messages
		MsgUserNotFound:       "Utilisateur non trouvé",
		MsgUserProfileUpdated: "Profil mis à jour avec succès",
		MsgUserDeleted:        "Compte utilisateur supprimé avec succès",

		// Manga messages
		MsgMangaCreated:         "Manga créé avec succès",
		MsgMangaUpdated:         "Manga mis à jour avec succès",
		MsgMangaDeleted:         "Manga supprimé avec succès",
		MsgMangaNotFound:        "Manga non trouvé",
		MsgMangaChapterAdded:    "Chapitre ajouté avec succès",
		MsgMangaChapterNotFound: "Chapitre non trouvé",

		// System messages
		MsgSystemInternalError:      "Erreur interne du serveur",
		MsgSystemServiceUnavailable: "Service temporairement indisponible",
		MsgSystemRateLimitExceeded:  "Trop de demandes. Veuillez réessayer plus tard",
		MsgSystemMaintenance:        "Système en maintenance",

		// Invite codes
		MsgInviteCodeUsed:      "Le code d'invitation a déjà été utilisé",
		MsgInviteCodeExpired:   "Le code d'invitation a expiré",
		MsgInviteCodeInvalid:   "Code d'invitation invalide",
		MsgInviteCodeGenerated: "Code d'invitation généré avec succès",
	}
}
