package routers

import (
	"github.com/Madou-Shinni/gin-quickstart/api/handle"
	"github.com/gin-gonic/gin"
)

var systemFileHandle = handle.NewSystemFileHandle()

// 注册路由
func SystemFileRouterRegister(r *gin.RouterGroup) {
	systemFileGroup := r.Group("systemFile")
	{
		systemFileGroup.POST("", systemFileHandle.Add)
		systemFileGroup.POST("/upload", systemFileHandle.Upload)
		systemFileGroup.GET("/download", systemFileHandle.Download)
		systemFileGroup.GET("/search", systemFileHandle.Search)
		systemFileGroup.DELETE("", systemFileHandle.Delete)
		systemFileGroup.DELETE("/delete-batch", systemFileHandle.DeleteByIds)
		systemFileGroup.GET("/:id", systemFileHandle.Find)
		systemFileGroup.GET("/list", systemFileHandle.List)
		systemFileGroup.PUT("", systemFileHandle.Update)
	}
}
