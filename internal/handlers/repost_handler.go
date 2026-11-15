package handlers

import (
	"net/http"

	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
)

type RepostHandler struct {
	service *servise.RepostService
}

func NewRepostHandler(service *servise.RepostService) *RepostHandler {
	return &RepostHandler{service: service}
}

func (h *RepostHandler) RegisterRoutes(r *gin.Engine) {
	reposts := r.Group("/api/reposts")
	reposts.GET("", h.GetAllForUser)
	reposts.POST("", h.Create)
}

func (h *RepostHandler) GetAllForUser(c *gin.Context) {
	var query struct {
		UserID uint `form:"user_id" binding:"required,min=1"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reposts, err := h.service.GetAllUserRepost(query.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reposts"})
		return
	}

	c.JSON(http.StatusOK, reposts)
}

func (h *RepostHandler) Create(c *gin.Context) {
	var req model.CreateRepostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repost, err := h.service.CreateRepost(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create repost"})
		return
	}

	c.JSON(http.StatusCreated, repost)
}
