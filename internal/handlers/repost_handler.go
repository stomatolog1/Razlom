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

func (h *RepostHandler) RegisterPublicRoutes(r gin.IRoutes) {
	r.GET("/api/users/:id/reposts", h.GetPublicProfile)
}

func (h *RepostHandler) GetAllForUser(c *gin.Context) {
	reposts, err := h.service.GetAllUserRepost(CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reposts"})
		return
	}

	c.JSON(http.StatusOK, reposts)
}

func (h *RepostHandler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.GET("/reposts", h.GetAllForUser)
	r.POST("/reposts", h.Create)
}

func (h *RepostHandler) GetPublicProfile(c *gin.Context) {
	userID, err := ParseUserIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reposts, err := h.service.GetAllUserRepost(userID)
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

	repost, err := h.service.CreateRepost(CurrentUserID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create repost"})
		return
	}

	c.JSON(http.StatusCreated, repost)
}
