package domain

import (
	"github.com/Madou-Shinni/gin-quickstart/pkg/request"
)

type SystemFile struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	FileName   string `json:"file_name"`
	Path       string `json:"path"`
	IsDir      bool   `json:"is_dir"`
	Size       int64  `json:"size"`       // 文件大小
	CreateTime string `json:"createTime"` // 创建时间
}

type PageSystemFileSearch struct {
	SystemFile
	request.PageSearch
}

func (SystemFile) TableName() string {
	return "system_file"
}
