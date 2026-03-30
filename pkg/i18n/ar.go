// pkg/i18n/ar.go
package i18n

// getArabicMessages returns Arabic message translations
func getArabicMessages() map[MessageKey]string {
	return map[MessageKey]string{
		// Authentication messages
		MsgAuthLoginSuccess:       "تم تسجيل الدخول بنجاح",
		MsgAuthLoginFailed:        "فشل في تسجيل الدخول",
		MsgAuthLogoutSuccess:      "تم تسجيل الخروج بنجاح",
		MsgAuthSignupSuccess:      "تم إنشاء الحساب بنجاح",
		MsgAuthUserExists:         "المستخدم موجود بالفعل",
		MsgAuthInvalidCredentials: "البريد الإلكتروني أو كلمة المرور غير صحيحة",
		MsgAuthTokenExpired:       "انتهت صلاحية رمز المصادقة",
		MsgAuthTokenInvalid:       "رمز المصادقة غير صحيح",
		MsgAuthAccessDenied:       "تم رفض الوصول",
		MsgAuthInsufficientRole:   "صلاحيات غير كافية لهذا الإجراء",

		// Validation messages
		MsgValidationRequired:      "هذا الحقل مطلوب",
		MsgValidationMinLength:     "الحد الأدنى للطول هو %d أحرف",
		MsgValidationMaxLength:     "الحد الأقصى للطول هو %d أحرف",
		MsgValidationInvalidFormat: "تنسيق غير صحيح",
		MsgValidationInvalidEmail:  "عنوان البريد الإلكتروني غير صحيح",
		MsgValidationPasswordWeak:  "كلمة المرور ضعيفة جداً. يجب أن تحتوي على الأقل 8 أحرف، حرف كبير واحد، حرف صغير واحد، رقم واحد، ورمز خاص واحد",

		// User messages
		MsgUserNotFound:       "المستخدم غير موجود",
		MsgUserProfileUpdated: "تم تحديث الملف الشخصي بنجاح",
		MsgUserDeleted:        "تم حذف حساب المستخدم بنجاح",

		// Manga messages
		MsgMangaCreated:         "تم إنشاء المانجا بنجاح",
		MsgMangaUpdated:         "تم تحديث المانجا بنجاح",
		MsgMangaDeleted:         "تم حذف المانجا بنجاح",
		MsgMangaNotFound:        "المانجا غير موجودة",
		MsgMangaChapterAdded:    "تم إضافة الفصل بنجاح",
		MsgMangaChapterNotFound: "الفصل غير موجود",

		// System messages
		MsgSystemInternalError:      "خطأ داخلي في الخادم",
		MsgSystemServiceUnavailable: "الخدمة غير متاحة مؤقتاً",
		MsgSystemRateLimitExceeded:  "طلبات كثيرة جداً. يرجى المحاولة لاحقاً",
		MsgSystemMaintenance:        "النظام قيد الصيانة",

		// Invite codes
		MsgInviteCodeUsed:      "رمز الدعوة مستخدم بالفعل",
		MsgInviteCodeExpired:   "انتهت صلاحية رمز الدعوة",
		MsgInviteCodeInvalid:   "رمز الدعوة غير صحيح",
		MsgInviteCodeGenerated: "تم إنشاء رمز الدعوة بنجاح",
	}
}
