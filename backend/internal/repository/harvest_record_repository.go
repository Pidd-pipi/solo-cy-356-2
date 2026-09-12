package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// HarvestRecordRepository 收成记录仓储接口。
type HarvestRecordRepository interface {
	Create(h *model.HarvestRecord) error
	CreateWithTx(tx *gorm.DB, h *model.HarvestRecord) error
	Update(h *model.HarvestRecord) error
	Delete(id uint) error
	FindByID(id uint) (*model.HarvestRecord, error)
	List(pq util.PageQuery, userID uint) ([]model.HarvestRecord, int64, error)
	ListByPlan(planID uint) ([]model.HarvestRecord, error)
	CountByUser(userID uint) (int64, error)
	SumWeightByYear(userID uint, year int) (float64, int, error)
	GroupByCropType(userID uint, year int) (map[string]float64, error)
	GroupByQuality(userID uint, year int) (map[string]int, error)
}

type harvestRecordRepository struct {
	db *gorm.DB
}

// NewHarvestRecordRepository 构造收成记录仓储。
func NewHarvestRecordRepository(db *gorm.DB) HarvestRecordRepository {
	return &harvestRecordRepository{db: db}
}

func (r *harvestRecordRepository) Create(h *model.HarvestRecord) error {
	return r.db.Create(h).Error
}

// CreateWithTx 在指定事务内创建收成记录。
func (r *harvestRecordRepository) CreateWithTx(tx *gorm.DB, h *model.HarvestRecord) error {
	return tx.Create(h).Error
}

func (r *harvestRecordRepository) Update(h *model.HarvestRecord) error {
	return r.db.Save(h).Error
}

func (r *harvestRecordRepository) Delete(id uint) error {
	return r.db.Delete(&model.HarvestRecord{}, id).Error
}

func (r *harvestRecordRepository) FindByID(id uint) (*model.HarvestRecord, error) {
	var h model.HarvestRecord
	if err := r.db.Preload("Plan").Preload("User").First(&h, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &h, nil
}

func (r *harvestRecordRepository) List(pq util.PageQuery, userID uint) ([]model.HarvestRecord, int64, error) {
	var records []model.HarvestRecord
	var total int64
	q := r.db.Model(&model.HarvestRecord{}).Preload("Plan").Preload("User")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("harvest_date DESC, id DESC"), pq).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *harvestRecordRepository) ListByPlan(planID uint) ([]model.HarvestRecord, error) {
	var records []model.HarvestRecord
	if err := r.db.Where("plan_id = ?", planID).Order("harvest_date ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *harvestRecordRepository) CountByUser(userID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&model.HarvestRecord{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *harvestRecordRepository) SumWeightByYear(userID uint, year int) (float64, int, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	type row struct {
		Total float64
		Cnt   int
	}
	var out row
	err := r.db.Model(&model.HarvestRecord{}).
		Select("COALESCE(sum(weight_kg),0) as total, count(*) as cnt").
		Where("user_id = ? AND harvest_date >= ? AND harvest_date < ?", userID, start, end).
		Scan(&out).Error
	return out.Total, out.Cnt, err
}

func (r *harvestRecordRepository) GroupByCropType(userID uint, year int) (map[string]float64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	type row struct {
		CropType string
		Total    float64
	}
	var rows []row
	err := r.db.Table("harvest_records").
		Select("plans.crop_type as crop_type, COALESCE(sum(harvest_records.weight_kg),0) as total").
		Joins("LEFT JOIN planting_plans plans ON plans.id = harvest_records.plan_id").
		Where("harvest_records.user_id = ? AND harvest_records.harvest_date >= ? AND harvest_records.harvest_date < ?", userID, start, end).
		Group("plans.crop_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(rows))
	for _, v := range rows {
		out[v.CropType] = v.Total
	}
	return out, nil
}

func (r *harvestRecordRepository) GroupByQuality(userID uint, year int) (map[string]int, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)
	type row struct {
		Quality string
		Cnt     int
	}
	var rows []row
	err := r.db.Model(&model.HarvestRecord{}).
		Select("quality, count(*) as cnt").
		Where("user_id = ? AND harvest_date >= ? AND harvest_date < ?", userID, start, end).
		Group("quality").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(rows))
	for _, v := range rows {
		out[v.Quality] = v.Cnt
	}
	return out, nil
}
