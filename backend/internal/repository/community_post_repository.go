package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// CommunityRepository 社区帖子仓储接口。
type CommunityRepository interface {
	Create(p *model.CommunityPost) error
	Update(p *model.CommunityPost) error
	Delete(id uint) error
	FindByID(id uint) (*model.CommunityPost, error)
	List(pq util.PageQuery, postType, status string) ([]model.CommunityPost, int64, error)
	IncrementLike(id uint) error
	CreateComment(c *model.CommunityComment) error
	ListComments(postID uint) ([]model.CommunityComment, error)
	CountByPostType() (map[string]int64, error)
}

type communityRepository struct {
	db *gorm.DB
}

// NewCommunityRepository 构造社区帖子仓储。
func NewCommunityRepository(db *gorm.DB) CommunityRepository {
	return &communityRepository{db: db}
}

func (r *communityRepository) Create(p *model.CommunityPost) error {
	return r.db.Create(p).Error
}

func (r *communityRepository) Update(p *model.CommunityPost) error {
	return r.db.Save(p).Error
}

func (r *communityRepository) Delete(id uint) error {
	return r.db.Delete(&model.CommunityPost{}, id).Error
}

func (r *communityRepository) FindByID(id uint) (*model.CommunityPost, error) {
	var p model.CommunityPost
	if err := r.db.Preload("User").First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *communityRepository) List(pq util.PageQuery, postType, status string) ([]model.CommunityPost, int64, error) {
	var posts []model.CommunityPost
	var total int64
	q := r.db.Model(&model.CommunityPost{}).Preload("User")
	if postType != "" {
		q = q.Where("post_type = ?", postType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := util.Paginate(q.Order("id DESC"), pq).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *communityRepository) IncrementLike(id uint) error {
	return r.db.Model(&model.CommunityPost{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *communityRepository) CreateComment(c *model.CommunityComment) error {
	return r.db.Create(c).Error
}

func (r *communityRepository) ListComments(postID uint) ([]model.CommunityComment, error) {
	var comments []model.CommunityComment
	if err := r.db.Preload("User").Where("post_id = ?", postID).Order("id ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *communityRepository) CountByPostType() (map[string]int64, error) {
	type row struct {
		PostType string
		Count    int64
	}
	var rows []row
	if err := r.db.Model(&model.CommunityPost{}).Select("post_type, count(*) as count").Group("post_type").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, v := range rows {
		out[v.PostType] = v.Count
	}
	return out, nil
}
