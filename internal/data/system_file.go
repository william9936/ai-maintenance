package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/Madou-Shinni/gin-quickstart/internal/domain"
	"github.com/Madou-Shinni/gin-quickstart/pkg/global"
	"github.com/Madou-Shinni/gin-quickstart/pkg/request"
	"github.com/Madou-Shinni/gin-quickstart/pkg/scopes"
)

type SystemFileRepo struct {
}

func (s *SystemFileRepo) Create(ctx context.Context, systemFile *domain.SystemFile) error {
	return global.DB.WithContext(ctx).Create(&systemFile).Error
}

func (s *SystemFileRepo) Delete(ctx context.Context, systemFile domain.SystemFile) error {
	return global.DB.WithContext(ctx).Delete(&systemFile).Error
}

func (s *SystemFileRepo) DeleteByIds(ctx context.Context, ids request.Ids) error {
	return global.DB.WithContext(ctx).Delete(&[]domain.SystemFile{}, ids.Ids).Error
}

func (s *SystemFileRepo) Update(ctx context.Context, systemFile domain.SystemFile) error {
	if systemFile.ID == 0 {
		return errors.New(fmt.Sprintf("missing %s.id", "systemFile"))
	}
	return nil
}

func (s *SystemFileRepo) Find(ctx context.Context, systemFile domain.SystemFile) (domain.SystemFile, error) {
	db := global.DB.WithContext(ctx).Model(&domain.SystemFile{})
	// TODO：条件过滤

	res := db.First(&systemFile)

	return systemFile, res.Error
}

func (s *SystemFileRepo) List(ctx context.Context, page domain.PageSystemFileSearch) ([]domain.SystemFile, int64, error) {
	var (
		systemFileList []domain.SystemFile
		count          int64
		err            error
	)
	// db
	db := global.DB.WithContext(ctx).Model(&domain.SystemFile{})

	// TODO：条件过滤

	err = db.Count(&count).Scopes(scopes.Paginate(page.PageSearch), scopes.OrderBy(page.OrderBy)).Find(&systemFileList).Error

	return systemFileList, count, err
}
