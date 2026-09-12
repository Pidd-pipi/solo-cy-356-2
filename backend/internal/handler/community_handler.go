package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// CommunityHandler 农友社区接口。
type CommunityHandler struct {
	communityService *service.CommunityService
}

// NewCommunityHandler 构造社区接口。
func NewCommunityHandler(communityService *service.CommunityService) *CommunityHandler {
	return &CommunityHandler{communityService: communityService}
}

// Create 发布帖子。
func (h *CommunityHandler) Create(c *gin.Context) {
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	p, err := h.communityService.Create(&req, util.GetUserID(c))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPostOutDTO(p))
}

// Update 更新帖子。
func (h *CommunityHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	p, err := h.communityService.Update(uint(id), claims.UserID, claims.Role, &req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPostOutDTO(p))
}

// Remove 删除帖子（软删除）。
func (h *CommunityHandler) Remove(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	if err := h.communityService.Remove(uint(id), claims.UserID, claims.Role); err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, gin.H{"removed": true})
}

// List 帖子分页列表。
func (h *CommunityHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	postType := c.Query("post_type")
	status := c.DefaultQuery("status", string(constants.PostStatusPublished))
	posts, total, err := h.communityService.List(pq, postType, status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.PostOutDTO, 0, len(posts))
	for i := range posts {
		list = append(list, dto.ToPostOutDTO(&posts[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// Get 帖子详情（含评论）。
func (h *CommunityHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	p, comments, err := h.communityService.GetByID(uint(id))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	out := dto.ToPostOutDTO(p)
	commentDTOs := make([]dto.CommunityCommentDTO, 0, len(comments))
	for i := range comments {
		commentDTOs = append(commentDTOs, *dto.ToCommunityCommentDTO(&comments[i]))
	}
	out.Comments = commentDTOs
	util.OK(c, out)
}

// Like 点赞帖子。
func (h *CommunityHandler) Like(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	if err := h.communityService.Like(uint(id), claims.UserID); err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, gin.H{"liked": true})
}

// Comment 评论帖子。
func (h *CommunityHandler) Comment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	comment, err := h.communityService.Comment(uint(id), claims.UserID, req.Content)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToCommunityCommentDTO(comment))
}
