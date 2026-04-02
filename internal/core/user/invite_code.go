// ----- START OF FILE: backend/MS-AI/internal/core/user/invite_code.go -----
// internal/core/invite_code.go
package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InviteCode يمثل رمز الدعوة الذي يُستخدم مرة واحدة للتسجيل.
type InviteCode struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Code      string             `bson:"code" json:"code" validate:"required"`
	IsUsed    bool               `bson:"is_used" json:"is_used"`           // لضمان الاستخدام مرة واحدة
	UsedBy    primitive.ObjectID `bson:"used_by,omitempty" json:"used_by"` // معرف المستخدم الذي استخدمه
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`     // يمكن إضافة تاريخ انتهاء صلاحية
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// CollectionName هو اسم Collection في MongoDB لهذا النموذج.
const InviteCodeCollectionName = "invite_codes"

// ----- END OF FILE: backend/MS-AI/internal/core/user/invite_code.go -----
