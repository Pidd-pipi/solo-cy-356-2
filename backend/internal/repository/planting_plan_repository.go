package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// PlantingPlanRepository 种植计划仓储接口。
type PlantingPlanRepository interface {
	Create(p *model.PlantingPlan) error
	CreateWithTx(tx *gorm.DB, p *model.PlantingPlan) error
	Update(p *model.PlantingPlan) error
	UpdateWithTx(tx *gorm.DB, p *model.PlantingPlan) error
	FindByID(id uint) (*model.PlantingPlan, error)
	List(pq util.PageQuery, userID uint, status string) ([]model.PlantingPlan, int64, error)
	ListByUser(userID uint, pq util.PageQuery) ([]model.PlantingPlan, int64, error)
	CountByUser(userID uint) (int64, error)
	CountByStatus() (map[string]int64, error)
	CountActiveByUser(userID uint) (int64, error)
}

type plantingPlanRepository struct {
	db *gorm.DB
}

// NewPlantingPlanRepository 构造种植计划仓储。
func NewPlantingPlanRepository(db *gorm.DB) PlantingPlanRepository {
	return &plantingPlanRepository{db: db}
}

func (r *plantingPlanRepository) Create(p *model.PlantingPlan) error {
	return r.db.Create(p).Error
}

// CreateWithTx 在指定事务内创建种植计划。
func (r *plantingPlanRepository) CreateWithTx(tx *gorm.DB, p *model.PlantingPlan) error {
	return tx.Create(p).Error
}

// UpdateWithTx 在指定事务内更新种植计划。
func (r *plantingPlanRepository) UpdateWithTx(tx *gorm.DB, p *model.PlantingPlan) error {
	return tx.Save(p).Error
}

func (r *plantingPlanRepository) Update(p *model.PlantingPlan) error {
	return r.db.Save(p).Error
}

func (r *plantingPlanRepository) FindByID(id uint) (*model.PlantingPlan, error) {
	var p model.PlantingPlan
	if err := r.db.Preload("Plot").Preload("User").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *plantingPlanRepository) List(pq util.PageQuery, userID uint, status string) ([]model.PlantingPlan, int64, error) {
	var plans []model.PlantingPlan
	var total int64
	q := r.db.Model(&model.PlantingPlan{}).Preload("Plot").Preload("User")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id DESC"), pq).Find(&plans).Error; err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}

// ListByUser 按用户分页列表（与 harvest 列表复用同一仓储方法族）。
func (r *plantingPlanRepository) ListByUser(userID uint, pq util.PageQuery) ([]model.PlantingPlan, int64, error) {
	return r.List(pq, userID, "")
}

func (r *plantingPlanRepository) CountByUser(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.PlantingPlan{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *plantingPlanRepository) CountByStatus() (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.Model(&model.PlantingPlan{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, v := range rows {
		out[v.Status] = v.Count
	}
	return out, nil
}

func (r *plantingPlanRepository) CountActiveByUser(userID uint) (int64, error) {
	var total int64
	err := r.db.Model(&model.PlantingPlan{}).
		Where("user_id = ? AND status IN ?", userID, []string{"planned", "planting", "growing", "harvesting"}).
		Count(&total).Error
	return total, err
}
