package handle

import (
	"errors"

	"github.com/Madou-Shinni/gin-quickstart/internal/domain"
	"github.com/Madou-Shinni/gin-quickstart/internal/service"
	"github.com/Madou-Shinni/gin-quickstart/pkg/constant"
	"github.com/Madou-Shinni/gin-quickstart/pkg/request"
	"github.com/Madou-Shinni/gin-quickstart/pkg/response"
	"github.com/Madou-Shinni/gin-quickstart/pkg/tools"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SystemFileHandle struct {
	s *service.SystemFileService
}

func NewSystemFileHandle() *SystemFileHandle {
	return &SystemFileHandle{s: service.NewSystemFileService()}
}

// Add 创建SystemFile
// @Tags     SystemFile
// @Summary  创建SystemFile
// @accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data body     service.AddSystemFileReq true "创建SystemFile"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile [post]
func (cl *SystemFileHandle) Add(c *gin.Context) {
	var req service.AddSystemFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) {
			response.Error(c, constant.CODE_INVALID_PARAMETER, tools.TransErrs(errs))
			return
		}
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	if systemFile, err := cl.s.Add(c.Request.Context(), req); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_ADD_FAILED, constant.CODE_ADD_FAILED.Msg())
		return
	} else {
		response.Success(c, systemFile)
	}
}

// Upload
// @Tags     SystemFile
// @Summary  上传SystemFile
// @accept   multipart/form-data
// @Produce  application/json
// @Param    filePath formData string true "上传SystemFile"
// @Param    files 	  formData file   true "files列表"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile/upload [post]
func (cl *SystemFileHandle) Upload(c *gin.Context) {
	var (
		err error
		req service.UploadSystemFileReq
	)
	// 获取上传目录
	form, _ := c.MultipartForm()
	fileHeaders := form.File["files"]
	filePath, ok := form.Value["filePath"]
	if !ok || len(filePath) == 0 {
		response.Error(c, constant.CODE_INVALID_PARAMETER, "filePath is required")
		return
	}

	req.FilePath = filePath[0]
	req.Files = fileHeaders

	// 上传文件
	if err = cl.s.Upload(c.Request.Context(), req); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_ADD_FAILED, constant.CODE_ADD_FAILED.Msg())
		return
	}

	response.Success(c, filePath)
}

// Download
// @Tags     SystemFile
// @Summary  下载SystemFile
// @accept   application/json
// @Produce  application/json
// @Param    filePath  query string true "文件路径"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile/download [post]
func (cl *SystemFileHandle) Download(c *gin.Context) {
	var (
		req service.SystemFilePathReq
	)

	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	if req.FilePath == "" {
		response.Error(c, constant.CODE_INVALID_PARAMETER, "filePath is required")
		return
	}

	c.File(req.FilePath)
}

// Search
// @Tags     SystemFile
// @Summary  搜索SystemFile
// @accept   application/json
// @Produce  application/json
// @Param    data  query service.SearchSystemFileReq true "搜索SystemFile"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile/search [get]
func (cl *SystemFileHandle) Search(c *gin.Context) {
	var (
		req service.SearchSystemFileReq
	)

	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	search, err := cl.s.Search(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_FIND_FAILED, constant.CODE_FIND_FAILED.Msg())
		return
	}

	response.Success(c, search)
}

// Delete 删除SystemFile
// @Tags     SystemFile
// @Summary  删除SystemFile
// @accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data body     service.RemoveSystemFileReq true "删除SystemFile"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile [delete]
func (cl *SystemFileHandle) Delete(c *gin.Context) {
	var systemFile service.RemoveSystemFileReq
	if err := c.ShouldBindJSON(&systemFile); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	if err := cl.s.Delete(c.Request.Context(), systemFile); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_DELETE_FAILED, constant.CODE_DELETE_FAILED.Msg())
		return
	}

	response.Success(c)
}

// DeleteByIds 批量删除SystemFile
// @Tags     SystemFile
// @Summary  批量删除SystemFile
// @accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data body     request.Ids true "批量删除SystemFile"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile/delete-batch [delete]
func (cl *SystemFileHandle) DeleteByIds(c *gin.Context) {
	var ids request.Ids
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	if err := cl.s.DeleteByIds(c.Request.Context(), ids); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_DELETE_FAILED, constant.CODE_DELETE_FAILED.Msg())
		return
	}

	response.Success(c)
}

// Update 修改SystemFile
// @Tags     SystemFile
// @Summary  修改SystemFile
// @accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data body     domain.SystemFile true "修改SystemFile"
// @Success  200  {string} string            "{"code":200,"msg":"","data":{}"}"
// @Router   /systemFile [put]
func (cl *SystemFileHandle) Update(c *gin.Context) {
	var systemFile domain.SystemFile
	if err := c.ShouldBindJSON(&systemFile); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	if err := cl.s.Update(c.Request.Context(), systemFile); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_UPDATE_FAILED, constant.CODE_UPDATE_FAILED.Msg())
		return
	}

	response.Success(c)
}

// Find 查询SystemFile
// @Tags     SystemFile
// @Summary  查询SystemFile
// @accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    id path     uint true "查询SystemFile"
// @Success  200  {string} string            "{"code":200,"msg":"查询成功","data":{}"}"
// @Router   /systemFile/{id} [get]
func (cl *SystemFileHandle) Find(c *gin.Context) {
	var systemFile domain.SystemFile
	if err := c.ShouldBindUri(&systemFile); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	res, err := cl.s.Find(c.Request.Context(), systemFile)

	if err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_FIND_FAILED, constant.CODE_FIND_FAILED.Msg())
		return
	}

	response.Success(c, res)
}

// List 查询SystemFile列表
// @Tags     SystemFile
// @Summary  查询SystemFile列表
// @accept   application/json
// @Produce  application/json
// @Security ApiKeyAuth
// @Param    data query     service.SystemFilePathReq true "查询SystemFile列表"
// @Success  200  {string} string            "{"code":200,"msg":"查询成功","data":{}"}"
// @Router   /systemFile/list [get]
func (cl *SystemFileHandle) List(c *gin.Context) {
	var systemFile service.SystemFilePathReq
	if err := c.ShouldBindQuery(&systemFile); err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_INVALID_PARAMETER, constant.CODE_INVALID_PARAMETER.Msg())
		return
	}

	res, err := cl.s.List(c.Request.Context(), systemFile)

	if err != nil {
		c.Error(err)
		response.Error(c, constant.CODE_FIND_FAILED, constant.CODE_FIND_FAILED.Msg())
		return
	}

	response.Success(c, res)
}
