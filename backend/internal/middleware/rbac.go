package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// RequireRoles RBAC 权限中间件：校验当前用户角色。
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := util.GetClaims(c)
		if !ok {
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgLoginRequired)
			c.Abort()
			return
		}
		for _, r := range roles {
			if claims.Role == r {
				c.Next()
				return
			}
		}
		util.Fail(c, http.StatusForbidden, constants.CodeForbidden, fmt.Sprintf(constants.MsgPermissionDenied, util.RoleText(claims.Role), util.RoleText(roles[0])))
		c.Abort()
	}
}
