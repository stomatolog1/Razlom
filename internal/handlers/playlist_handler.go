package handlers

import (
	"net/http"

	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
)

type PlaylistHandler struct {
	service *servise.PlaylistService
}

func NewPlaylistHandler(service *servise.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{service: service}
}

func (h *PlaylistHandler) RegisterRoutes(r *gin.Engine) {
	playlists := r.Group("/api/playlists")
	playlists.GET("", h.GetPublicForUser)
	playlists.POST("", h.Create)
}

func (h *PlaylistHandler) GetPublicForUser(c *gin.Context) {
	var query struct {
		UserID uint `form:"user_id" binding:"required,min=1"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlists, err := h.service.GetPublicPlaylist(query.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get playlists"})
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *PlaylistHandler) Create(c *gin.Context) {
	var req model.NewPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.service.CreateNewPlaylist(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}
