package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// AuditRepository 审计日志仓储接口。
type AuditRepository interface {
	Create(a *model.AuditLog) error
	List(pq util.PageQuery, action, username string) ([]model.AuditLog, int64, error)
	FindByID(id uint) (*model.AuditLog, error)
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 构造审计日志仓储。
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(a *model.AuditLog) error {
	return r.db.Create(a).Error
}

func (r *auditRepository) List(pq util.PageQuery, action, username string) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	q := r.db.Model(&model.AuditLog{})
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if username != "" {
		q = q.Where("username = ?", username)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id DESC"), pq).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *auditRepository) FindByID(id uint) (*model.AuditLog, error) {
	var a model.AuditLog
	if err := r.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}
