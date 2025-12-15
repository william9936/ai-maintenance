package handle

import (
	"time"

	"github.com/Madou-Shinni/gin-quickstart/internal/domain"
	"github.com/Madou-Shinni/gin-quickstart/internal/service"
	"github.com/Madou-Shinni/gin-quickstart/pkg/constant"
	"github.com/Madou-Shinni/gin-quickstart/pkg/global"
	"github.com/Madou-Shinni/gin-quickstart/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	pwd           = "admin"
	hashPwd, _    = bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	casbinService = service.NewSysCasbinService()
	defaultRole   = domain.SysRole{
		ParentID: 0,
		RoleName: "超级管理员",
	}
	defaultUser = domain.SysUser{
		Account:  "admin",
		Password: string(hashPwd),
		NickName: "超级管理员",
		Roles: []domain.SysRole{
			defaultRole,
		},
	}
)

type SystemHandle struct {
	casbinService  *service.SysCasbinService
	monitorService *service.MonitorService
}

func NewSystemHandle() *SystemHandle {
	return &SystemHandle{
		casbinService:  casbinService,
		monitorService: service.MonitorServiceEx,
	}
}

// Init 系统初始化
// @Tags     System
// @Summary  系统初始化
// @accept   application/json
// @Produce  application/json
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /system/init [post]
func (cl *SystemHandle) Init(c *gin.Context) {
	var err error
	var count int64
	global.DB.WithContext(c.Request.Context()).Model(&domain.SysRole{}).Count(&count)
	if count > 0 {
		response.Error(c, constant.CODE_ADD_FAILED, "已完成初始化，请勿重复")
		return
	}
	err = global.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		// 添加管理员
		err = tx.Model(&domain.SysUser{}).Create(&defaultUser).Error
		if err != nil {
			return err
		}
		// 添加管理员默认角色
		defaultUser.DefaultRole = defaultUser.Roles[0].ID
		err = tx.Updates(&defaultUser).Error
		if err != nil {
			return err
		}
		// 添加管理员权限
		err = casbinService.AddUserRoles(c.Request.Context(), domain.UserRolesReq{
			UserID: defaultUser.ID,
			Roles:  []uint{defaultUser.DefaultRole},
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		response.Error(c, constant.CODE_ADD_FAILED, err.Error())
		return
	}

	defaultUser.Password = pwd
	response.Success(c, defaultUser)
}

// MonitorState 监控状态
// @Tags     System
// @Summary  监控状态
// @accept   application/json
// @Produce  application/json
// @Success  200  {string} string            "{"code":200,"msg":"查询成功","data":{}"}"
// @Router   /system/monitor [get]
func (cl *SystemHandle) MonitorState(c *gin.Context) {
	var req service.MonitorStateReq

	req.StartTime = time.Now().Add(-time.Hour * 24).Unix()
	req.EndTime = time.Now().Unix()
	resp, err := cl.monitorService.All(c.Request.Context(), req.StartTime, req.EndTime)
	if err != nil {
		response.Error(c, constant.CODE_FIND_FAILED, err.Error())
		return
	}

	response.Success(c, resp)
}
