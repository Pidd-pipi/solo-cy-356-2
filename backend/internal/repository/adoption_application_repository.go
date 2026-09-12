package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// AdoptionApplicationRepository 认养申请仓储接口。
type AdoptionApplicationRepository interface {
	CreateWithTx(tx *gorm.DB, a *model.AdoptionApplication) error
	UpdateWithTx(tx *gorm.DB, a *model.AdoptionApplication) error
	FindByID(id uint) (*model.AdoptionApplication, error)
	FindByIDForUpdate(tx *gorm.DB, id uint) (*model.AdoptionApplication, error)
	LockApplicant(tx *gorm.DB, userID uint) error
	CountActiveByUser(tx *gorm.DB, userID uint) (int64, error)
	CountPendingByPlot(tx *gorm.DB, plotID uint) (int64, error)
	FindEarliestWaitlistedForUpdate(tx *gorm.DB, plotID uint) (*model.AdoptionApplication, error)
	ListByUser(pq util.PageQuery, userID uint) ([]model.AdoptionApplication, int64, error)
	List(pq util.PageQuery, status string) ([]model.AdoptionApplication, int64, error)
}

type adoptionApplicationRepository struct {
	db *gorm.DB
}

// NewAdoptionApplicationRepository 构造认养申请仓储。
func NewAdoptionApplicationRepository(db *gorm.DB) AdoptionApplicationRepository {
	return &adoptionApplicationRepository{db: db}
}

func (r *adoptionApplicationRepository) CreateWithTx(tx *gorm.DB, a *model.AdoptionApplication) error {
	return tx.Create(a).Error
}

func (r *adoptionApplicationRepository) UpdateWithTx(tx *gorm.DB, a *model.AdoptionApplication) error {
	return tx.Save(a).Error
}

func (r *adoptionApplicationRepository) FindByID(id uint) (*model.AdoptionApplication, error) {
	var a model.AdoptionApplication
	if err := r.db.Preload("Plot").Preload("User").First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// FindByIDForUpdate 审核/撤回使用 SELECT ... FOR UPDATE 行锁（事务内执行）。
func (r *adoptionApplicationRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*model.AdoptionApplication, error) {
	var a model.AdoptionApplication
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Plot").Preload("User").First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// LockApplicant 锁定申请人用户行（事务内执行），串行化同一用户的并发申请，
// 保证“同一时间仅一份进行中申请”的约束在并发下依然成立。
func (r *adoptionApplicationRepository) LockApplicant(tx *gorm.DB, userID uint) error {
	var u model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// CountActiveByUser 统计用户进行中（待审核 + 候补中）的申请数。
func (r *adoptionApplicationRepository) CountActiveByUser(tx *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := tx.Model(&model.AdoptionApplication{}).
		Where("user_id = ? AND status IN ?", userID, []string{string(constants.ApplicationPending), string(constants.ApplicationWaitlisted)}).
		Count(&count).Error
	return count, err
}

// CountPendingByPlot 统计地块当前待审核申请数（决定是否进入候补队列）。
func (r *adoptionApplicationRepository) CountPendingByPlot(tx *gorm.DB, plotID uint) (int64, error) {
	var count int64
	err := tx.Model(&model.AdoptionApplication{}).
		Where("plot_id = ? AND status = ?", plotID, string(constants.ApplicationPending)).
		Count(&count).Error
	return count, err
}

// FindEarliestWaitlistedForUpdate 查询地块最早的候补申请（按申请时间升序，行锁）。
func (r *adoptionApplicationRepository) FindEarliestWaitlistedForUpdate(tx *gorm.DB, plotID uint) (*model.AdoptionApplication, error) {
	var a model.AdoptionApplication
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("plot_id = ? AND status = ?", plotID, string(constants.ApplicationWaitlisted)).
		Order("created_at ASC, id ASC").
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// ListByUser 我的申请分页列表（进度查询）。
func (r *adoptionApplicationRepository) ListByUser(pq util.PageQuery, userID uint) ([]model.AdoptionApplication, int64, error) {
	var apps []model.AdoptionApplication
	var total int64
	q := r.db.Model(&model.AdoptionApplication{}).Preload("Plot").Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id DESC"), pq).Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

// List 全部申请分页列表（管理员审核，可按状态过滤）。
func (r *adoptionApplicationRepository) List(pq util.PageQuery, status string) ([]model.AdoptionApplication, int64, error) {
	var apps []model.AdoptionApplication
	var total int64
	q := r.db.Model(&model.AdoptionApplication{}).Preload("Plot").Preload("User")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id DESC"), pq).Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}
