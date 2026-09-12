package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// PlotRepository 地块仓储接口。
type PlotRepository interface {
	Create(p *model.Plot) error
	Update(p *model.Plot) error
	FindByID(id uint) (*model.Plot, error)
	FindByIDForUpdate(tx *gorm.DB, id uint) (*model.Plot, error)
	UpdateWithTx(tx *gorm.DB, p *model.Plot) error
	FindByCode(code string) (*model.Plot, error)
	List(pq util.PageQuery, status string) ([]model.Plot, int64, error)
	CountByStatus() (map[string]int64, error)
}

type plotRepository struct {
	db *gorm.DB
}

// NewPlotRepository 构造地块仓储。
func NewPlotRepository(db *gorm.DB) PlotRepository {
	return &plotRepository{db: db}
}

func (r *plotRepository) Create(p *model.Plot) error {
	return r.db.Create(p).Error
}

func (r *plotRepository) Update(p *model.Plot) error {
	return r.db.Save(p).Error
}

func (r *plotRepository) FindByID(id uint) (*model.Plot, error) {
	var p model.Plot
	if err := r.db.Preload("Adopter").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// FindByIDForUpdate 并发认养使用 SELECT ... FOR UPDATE 行锁（事务内执行）。
func (r *plotRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*model.Plot, error) {
	var p model.Plot
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// UpdateWithTx 在指定事务内更新地块。
func (r *plotRepository) UpdateWithTx(tx *gorm.DB, p *model.Plot) error {
	return tx.Save(p).Error
}

func (r *plotRepository) FindByCode(code string) (*model.Plot, error) {
	var p model.Plot
	if err := r.db.Where("code = ?", code).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *plotRepository) List(pq util.PageQuery, status string) ([]model.Plot, int64, error) {
	var plots []model.Plot
	var total int64
	q := r.db.Model(&model.Plot{}).Preload("Adopter")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id ASC"), pq).Find(&plots).Error; err != nil {
		return nil, 0, err
	}
	return plots, total, nil
}

func (r *plotRepository) CountByStatus() (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.Model(&model.Plot{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, v := range rows {
		out[v.Status] = v.Count
	}
	return out, nil
}
