package handlers

import (
	"net/http"

	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
	service *servise.LikeService
}

func NewLikeHandler(service *servise.LikeService) *LikeHandler {
	return &LikeHandler{service: service}
}

func (h *LikeHandler) RegisterRoutes(r *gin.Engine) {
	likes := r.Group("/api/likes")
	likes.GET("", h.GetAllForUser)
	likes.POST("", h.Create)
}

func (h *LikeHandler) GetAllForUser(c *gin.Context) {
	var query struct {
		UserID uint `form:"user_id" binding:"required,min=1"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	likes, err := h.service.GetAllForUser(query.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get likes"})
		return
	}

	c.JSON(http.StatusOK, likes)
}

func (h *LikeHandler) Create(c *gin.Context) {
	var req model.NewLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	like, err := h.service.CreateLike(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create like"})
		return
	}

	c.JSON(http.StatusCreated, like)
}
