package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// ErrNotFound 哨兵错误：资源不存在。
var ErrNotFound = errors.New("record not found")

// UserRepository 用户仓储接口。
type UserRepository interface {
	Create(u *model.User) error
	Update(u *model.User) error
	Delete(id uint) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	List(pq util.PageQuery, role, status string) ([]model.User, int64, error)
	CountByRole() (map[string]int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) Update(u *model.User) error {
	return r.db.Save(u).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) List(pq util.PageQuery, role, status string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.Model(&model.User{})
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id DESC"), pq).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) CountByRole() (map[string]int64, error) {
	type row struct {
		Role  string
		Count int64
	}
	var rows []row
	if err := r.db.Model(&model.User{}).Select("role, count(*) as count").Group("role").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, v := range rows {
		out[v.Role] = v.Count
	}
	return out, nil
}
