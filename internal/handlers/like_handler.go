package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LikeHandler struct {
	service *servise.LikeService
}

func NewLikeHandler(service *servise.LikeService) *LikeHandler {
	return &LikeHandler{service: service}
}

func (h *LikeHandler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.GET("/likes", h.GetAllForUser)
	r.POST("/likes", h.Create)
	r.DELETE("/likes/:musicID", h.Delete)
}

func (h *LikeHandler) GetAllForUser(c *gin.Context) {
	likes, err := h.service.GetAllForUser(CurrentUserID(c))
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

	like, err := h.service.CreateLike(CurrentUserID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create like"})
		return
	}

	c.JSON(http.StatusCreated, like)
}

func (h *LikeHandler) Delete(c *gin.Context) {
	musicID, err := strconv.Atoi(c.Param("musicID"))
	if err != nil || musicID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid music id"})
		return
	}
	if err := h.service.DeleteLike(CurrentUserID(c), musicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "like not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove like"})
		return
	}
	c.Status(http.StatusNoContent)
}
