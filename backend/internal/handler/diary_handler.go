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

// DiaryHandler 种植日记接口。
type DiaryHandler struct {
	diaryService *service.DiaryService
}

// NewDiaryHandler 构造种植日记接口。
func NewDiaryHandler(diaryService *service.DiaryService) *DiaryHandler {
	return &DiaryHandler{diaryService: diaryService}
}

// Create 发布日记。
func (h *DiaryHandler) Create(c *gin.Context) {
	var req dto.CreateDiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	d, err := h.diaryService.Create(&req, util.GetUserID(c))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToDiaryOutDTO(d))
}

// Update 更新日记。
func (h *DiaryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.UpdateDiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	d, err := h.diaryService.Update(uint(id), claims.UserID, claims.Role, &req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToDiaryOutDTO(d))
}

// Delete 删除日记。
func (h *DiaryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	if err := h.diaryService.Delete(uint(id), claims.UserID, claims.Role); err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, gin.H{"deleted": true})
}

// List 日记分页列表。
func (h *DiaryHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	claims, _ := util.GetClaims(c)
	var userID uint
	if claims.Role != string(constants.RoleAdmin) {
		userID = claims.UserID
	}
	planID, _ := strconv.ParseUint(c.Query("plan_id"), 10, 64)
	diaries, total, err := h.diaryService.List(pq, userID, uint(planID))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.DiaryOutDTO, 0, len(diaries))
	for i := range diaries {
		list = append(list, dto.ToDiaryOutDTO(&diaries[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// Get 日记详情（含评论）。
func (h *DiaryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	d, comments, err := h.diaryService.GetByID(uint(id))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	out := dto.ToDiaryOutDTO(d)
	commentDTOs := make([]dto.DiaryCommentDTO, 0, len(comments))
	for i := range comments {
		commentDTOs = append(commentDTOs, *dto.ToDiaryCommentDTO(&comments[i]))
	}
	out.Comments = commentDTOs
	util.OK(c, out)
}

// Like 点赞日记。
func (h *DiaryHandler) Like(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	if err := h.diaryService.Like(uint(id), claims.UserID); err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, gin.H{"liked": true})
}

// Comment 评论日记。
func (h *DiaryHandler) Comment(c *gin.Context) {
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
	comment, err := h.diaryService.Comment(uint(id), claims.UserID, req.Content)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToDiaryCommentDTO(comment))
}
