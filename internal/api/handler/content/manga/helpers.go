package manga

import (
	"github.com/835-droid/ms-ai-backend/internal/api/middleware"
	core "github.com/835-droid/ms-ai-backend/internal/core/common"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// getCallerInfo extracts the caller ID and roles from the gin context
func getCallerInfo(c *gin.Context) (primitive.ObjectID, []string, error) {
	uid, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return primitive.NilObjectID, nil, core.ErrUnauthorized
	}
	uidStr, _ := uid.(string)
	callerID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		return primitive.NilObjectID, nil, core.ErrUnauthorized
	}

	var roles []string
	if v, ok := c.Get(middleware.ContextUserRolesKey); ok {
		if r, ok := v.([]string); ok {
			roles = r
		}
	}

	return callerID, roles, nil
}
