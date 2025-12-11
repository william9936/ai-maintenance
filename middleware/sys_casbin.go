package middleware

import (
	"github.com/Madou-Shinni/gin-quickstart/common"
	"github.com/Madou-Shinni/gin-quickstart/internal/service"
	"github.com/Madou-Shinni/gin-quickstart/pkg/constant"
	"github.com/Madou-Shinni/gin-quickstart/pkg/response"
	"github.com/Madou-Shinni/go-logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CasbinHandler 拦截器
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := common.GetRoleIdFromCtx(c)
		//获取请求的PATH
		path := c.Request.URL.Path
		obj := path
		// 获取请求方法
		act := c.Request.Method
		// 获取用户的角色
		sub := constant.GetCasbinRoleKey(id)
		success, _ := service.Casbin().Enforce(sub, obj, act)
		logger.Info("CasbinHandler", zap.Any("sub", sub), zap.Any("obj", obj), zap.Any("act", act), zap.Any("success", success))
		if !success {
			response.Error(c, constant.CODE_NO_PERMISSIONS, constant.CODE_NO_PERMISSIONS.Msg())
			c.Abort()
			return
		}
		c.Next()
	}
}
