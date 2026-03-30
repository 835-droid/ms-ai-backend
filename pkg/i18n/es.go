// pkg/i18n/es.go
package i18n

// getSpanishMessages returns Spanish message translations
func getSpanishMessages() map[MessageKey]string {
	return map[MessageKey]string{
		// Authentication messages
		MsgAuthLoginSuccess:       "Inicio de sesión exitoso",
		MsgAuthLoginFailed:        "Error en el inicio de sesión",
		MsgAuthLogoutSuccess:      "Cierre de sesión exitoso",
		MsgAuthSignupSuccess:      "Cuenta creada exitosamente",
		MsgAuthUserExists:         "El usuario ya existe",
		MsgAuthInvalidCredentials: "Email o contraseña inválidos",
		MsgAuthTokenExpired:       "El token de autenticación ha expirado",
		MsgAuthTokenInvalid:       "Token de autenticación inválido",
		MsgAuthAccessDenied:       "Acceso denegado",
		MsgAuthInsufficientRole:   "Permisos insuficientes para esta acción",

		// Validation messages
		MsgValidationRequired:      "Este campo es obligatorio",
		MsgValidationMinLength:     "La longitud mínima es de %d caracteres",
		MsgValidationMaxLength:     "La longitud máxima es de %d caracteres",
		MsgValidationInvalidFormat: "Formato inválido",
		MsgValidationInvalidEmail:  "Dirección de email inválida",
		MsgValidationPasswordWeak:  "La contraseña es demasiado débil. Debe contener al menos 8 caracteres, una letra mayúscula, una letra minúscula, un número y un carácter especial",

		// User messages
		MsgUserNotFound:       "Usuario no encontrado",
		MsgUserProfileUpdated: "Perfil actualizado exitosamente",
		MsgUserDeleted:        "Cuenta de usuario eliminada exitosamente",

		// Manga messages
		MsgMangaCreated:         "Manga creado exitosamente",
		MsgMangaUpdated:         "Manga actualizado exitosamente",
		MsgMangaDeleted:         "Manga eliminado exitosamente",
		MsgMangaNotFound:        "Manga no encontrado",
		MsgMangaChapterAdded:    "Capítulo agregado exitosamente",
		MsgMangaChapterNotFound: "Capítulo no encontrado",

		// System messages
		MsgSystemInternalError:      "Error interno del servidor",
		MsgSystemServiceUnavailable: "Servicio temporalmente no disponible",
		MsgSystemRateLimitExceeded:  "Demasiadas solicitudes. Por favor, inténtelo de nuevo más tarde",
		MsgSystemMaintenance:        "Sistema en mantenimiento",

		// Invite codes
		MsgInviteCodeUsed:      "El código de invitación ya ha sido utilizado",
		MsgInviteCodeExpired:   "El código de invitación ha expirado",
		MsgInviteCodeInvalid:   "Código de invitación inválido",
		MsgInviteCodeGenerated: "Código de invitación generado exitosamente",
	}
}
