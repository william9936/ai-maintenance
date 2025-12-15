package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Madou-Shinni/gin-quickstart/internal/data"
	"github.com/Madou-Shinni/gin-quickstart/internal/domain"
	"github.com/Madou-Shinni/gin-quickstart/pkg/request"
	"github.com/Madou-Shinni/gin-quickstart/pkg/response"
	"github.com/Madou-Shinni/go-logger"
	"go.uber.org/zap"
)

type SystemFilePathReq struct {
	FilePath string `json:"filePath" form:"filePath"`
}

type AddSystemFileReq struct {
	FilePath string `json:"filePath" form:"filePath"`
	IsDir    bool   `json:"isDir" form:"isDir"`
}

type RemoveSystemFileReq struct {
	FilePath string `json:"filePath" form:"filePath"`
}

type UploadSystemFileReq struct {
	FilePath string                  `json:"filePath" form:"filePath"`
	Files    []*multipart.FileHeader `json:"files" form:"files"`
}

type SearchSystemFileReq struct {
	FilePath string `json:"filePath" form:"filePath"`
	FileName string `json:"fileName" form:"fileName"`
}

// 定义接口
type SystemFileRepo interface {
	Create(ctx context.Context, systemFile *domain.SystemFile) error
	Delete(ctx context.Context, systemFile domain.SystemFile) error
	Update(ctx context.Context, systemFile domain.SystemFile) error
	Find(ctx context.Context, systemFile domain.SystemFile) (domain.SystemFile, error)
	List(ctx context.Context, page domain.PageSystemFileSearch) ([]domain.SystemFile, int64, error)
	DeleteByIds(ctx context.Context, ids request.Ids) error
}

type SystemFileService struct {
	repo SystemFileRepo
}

func NewSystemFileService() *SystemFileService {
	return &SystemFileService{repo: &data.SystemFileRepo{}}
}

// Add 添加SystemFile
// 1.根据传递的参数创建文件或目录
// 2.验证传递的参数是否有效，例如路径是否存在等
func (s *SystemFileService) Add(ctx context.Context, req AddSystemFileReq) (domain.SystemFile, error) {
	// 验证参数有效性
	if req.FilePath == "" {
		return domain.SystemFile{}, errors.New("文件路径不能为空")
	}

	// 检查父目录是否存在
	parentDir := filepath.Dir(req.FilePath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		return domain.SystemFile{}, errors.New("父目录不存在")
	}

	// 根据类型创建文件或目录
	var info os.FileInfo
	var err error
	if !req.IsDir {
		// 创建文件
		file, err := os.Create(req.FilePath)
		if err != nil {
			logger.Error("os.Create(filePath)", zap.Error(err), zap.String("filePath", req.FilePath))
			return domain.SystemFile{}, err
		}
		defer file.Close()

		// 获取文件信息
		info, err = file.Stat()
		if err != nil {
			logger.Error("file.Stat()", zap.Error(err), zap.String("filePath", req.FilePath))
			return domain.SystemFile{}, err
		}
	} else {
		// 创建目录
		err = os.Mkdir(req.FilePath, 0755)
		if err != nil {
			logger.Error("os.Mkdir(filePath)", zap.Error(err), zap.String("filePath", req.FilePath))
			return domain.SystemFile{}, err
		}

		// 获取目录信息
		info, err = os.Stat(req.FilePath)
		if err != nil {
			logger.Error("os.Stat(filePath)", zap.Error(err), zap.String("filePath", req.FilePath))
			return domain.SystemFile{}, err
		}
	}

	// 构造SystemFile对象
	systemFile := domain.SystemFile{
		FileName:   filepath.Base(req.FilePath),
		Path:       req.FilePath,
		IsDir:      info.IsDir(),
		Size:       info.Size(),
		CreateTime: formatTime(info.ModTime()),
	}

	return systemFile, nil
}

// Delete 删除SystemFile
// 1.根据传递的参数删除文件或目录
// 2.验证传递的参数是否有效，例如路径是否存在等
func (s *SystemFileService) Delete(ctx context.Context, systemFile RemoveSystemFileReq) error {
	// 验证参数有效性
	if systemFile.FilePath == "" {
		return errors.New("文件路径不能为空")
	}

	// 检查文件或目录是否存在
	if _, err := os.Stat(systemFile.FilePath); os.IsNotExist(err) {
		return errors.New("文件或目录不存在")
	}

	// 检查是否为目录
	info, err := os.Stat(systemFile.FilePath)
	if err != nil {
		logger.Error("os.Stat(filePath)", zap.Error(err), zap.String("filePath", systemFile.FilePath))
		return err
	}
	if info.IsDir() {
		// 删除目录
		err := os.RemoveAll(systemFile.FilePath)
		if err != nil {
			logger.Error("os.RemoveAll(filePath)", zap.Error(err), zap.String("filePath", systemFile.FilePath))
			return err
		}
	} else {
		// 删除文件
		err := os.Remove(systemFile.FilePath)
		if err != nil {
			logger.Error("os.Remove(filePath)", zap.Error(err), zap.String("filePath", systemFile.FilePath))
			return err
		}
	}

	return nil
}

// Upload 上传SystemFile
// 1.根据传递的参数上传文件或目录
// 2.验证传递的参数是否有效，例如路径是否存在等
// 3.文件存在时，则根据mac系统默认处理方式处理，添加文件副本，添加1、2等
func (s *SystemFileService) Upload(ctx context.Context, req UploadSystemFileReq) error {
	// 验证参数有效性
	if req.FilePath == "" {
		return errors.New("文件路径不能为空")
	}

	if req.Files == nil {
		return errors.New("上传文件不能为空")
	}

	// 检查目标目录是否存在
	uploadDir := filepath.Dir(req.FilePath)
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return errors.New("目标目录不存在")
	}

	// 打开上传的文件
	for _, fileHeader := range req.Files {
		// 打开上传的文件
		src, err := fileHeader.Open()
		if err != nil {
			logger.Error("fileHeader.Open()", zap.Error(err), zap.String("fileName", fileHeader.Filename))
			return err
		}
		defer src.Close()

		// 处理文件存在的情况（按照mac系统默认方式添加副本）
		path := filepath.Join(uploadDir, fileHeader.Filename)
		dstPath := getDestinationPath(path)
		// 创建目标文件
		dst, err := os.Create(dstPath)
		if err != nil {
			logger.Error("os.Create(dstPath)", zap.Error(err), zap.String("dstPath", dstPath))
			return err
		}
		defer dst.Close()

		// 将上传的文件内容复制到目标文件
		if _, err = io.Copy(dst, src); err != nil {
			logger.Error("io.Copy(dst, src)", zap.Error(err), zap.String("dstPath", dstPath))
			return err
		}
	}

	return nil
}

// getDestinationPath 获取文件的目标路径，如果文件已存在，则按照mac系统默认方式添加副本
func getDestinationPath(filePath string) string {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	// 文件已存在，按照mac系统默认方式处理
	dir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// 查找可用的副本名称
	counter := 1
	for {
		// 构造副本名称，如：filename_1.txt
		newName := fmt.Sprintf("%s_副本%d%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newName)

		// 检查副本是否存在
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}

		counter++
	}
}

func (s *SystemFileService) Update(ctx context.Context, systemFile domain.SystemFile) error {
	if err := s.repo.Update(ctx, systemFile); err != nil {
		logger.Error("s.repo.Update(systemFile)", zap.Error(err), zap.Any("domain.SystemFile", systemFile))
		return err
	}

	return nil
}

func (s *SystemFileService) Find(ctx context.Context, systemFile domain.SystemFile) (domain.SystemFile, error) {
	res, err := s.repo.Find(ctx, systemFile)

	if err != nil {
		logger.Error("s.repo.Find(systemFile)", zap.Error(err), zap.Any("domain.SystemFile", systemFile))
		return res, err
	}

	return res, nil
}

// 1.根据传递的filePath查询系统文件列表（包含目录和文件
// 2.如果路径不存在，返回错误，路径为空，默认查询根目录
// 3.根目录需根据系统类型查询，Windows为所有盘符，Linux为根目录
// 输出：文件列表（包含目录和文件）文件名或目录名、文件大小、文件类型、创建时间、更新时间
func (s *SystemFileService) List(ctx context.Context, req SystemFilePathReq) (response.PageResponse, error) {
	var (
		pageRes response.PageResponse
	)

	// 获取查询路径
	filePath := req.FilePath

	// 如果路径为空，根据操作系统类型获取根目录
	if filePath == "" {
		filePath = getRootPath()
	}

	// 读取目录内容
	entries, err := os.ReadDir(filePath)
	if err != nil {
		logger.Error("os.ReadDir(filePath)", zap.Error(err), zap.String("filePath", filePath))
		return pageRes, err
	}

	// 构造文件列表
	fileList := make([]domain.SystemFile, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			logger.Error("entry.Info()", zap.Error(err), zap.String("entryName", entry.Name()))
			continue
		}

		// 文件类型 1.文件 2.目录
		var isDir bool
		if info.IsDir() {
			isDir = true
		}

		fileList = append(fileList, domain.SystemFile{
			FileName:   entry.Name(),
			Path:       filepath.Join(filePath, entry.Name()),
			IsDir:      isDir,
			Size:       info.Size(),
			CreateTime: formatTime(info.ModTime()), // Go标准库没有直接获取创建时间的跨平台方法
		})
	}

	// 设置分页响应
	pageRes.List = fileList
	pageRes.Total = int64(len(fileList))

	return pageRes, nil
}

// getRootPath 根据操作系统类型获取根目录
func getRootPath() string {
	// 获取操作系统类型
	osType := runtime.GOOS

	// 根据操作系统类型返回根目录
	switch osType {
	case "windows":
		// Windows系统返回所有盘符
		return ""
	case "linux", "darwin":
		// Linux和macOS系统返回根目录
		return "/"
	default:
		// 默认返回根目录
		return "/"
	}
}

// formatTime 格式化时间
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func (s *SystemFileService) DeleteByIds(ctx context.Context, ids request.Ids) error {
	if err := s.repo.DeleteByIds(ctx, ids); err != nil {
		logger.Error("s.DeleteByIds(ids)", zap.Error(err), zap.Any("ids request.Ids", ids))
		return err
	}

	return nil
}

// Search 根据文件路径和文件名搜索系统文件
// 1.根据传递的filePath查询fileName文件
// 2.支持子目录搜索
// 输出：文件列表（包含目录和文件）文件名或目录名、文件大小、文件类型、创建时间、更新时间
func (s *SystemFileService) Search(ctx context.Context, req SearchSystemFileReq) (response.PageResponse, error) {
	var (
		pageRes response.PageResponse
	)

	// 检查搜索路径是否存在
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		return pageRes, errors.New("搜索路径不存在")
	}

	// 搜索文件
	var matchedFiles []domain.SystemFile
	err := filepath.Walk(req.FilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error("filepath.Walk", zap.Error(err), zap.String("path", path))
			return nil // 继续遍历其他文件
		}

		// 检查文件名是否匹配
		if strings.Contains(info.Name(), req.FileName) {
			matchedFiles = append(matchedFiles, domain.SystemFile{
				FileName:   info.Name(),
				Path:       path,
				IsDir:      info.IsDir(),
				Size:       info.Size(),
				CreateTime: formatTime(info.ModTime()),
			})
		}

		return nil
	})

	if err != nil {
		logger.Error("搜索文件失败", zap.Error(err), zap.String("filePath", req.FilePath), zap.String("fileName", req.FileName))
		return pageRes, err
	}

	// 设置分页响应
	pageRes.List = matchedFiles
	pageRes.Total = int64(len(matchedFiles))

	return pageRes, nil
}
