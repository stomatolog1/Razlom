package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlaylistHandler struct {
	service   *servise.PlaylistService
	deletions *repository.DeletionRepository
}

func NewPlaylistHandler(service *servise.PlaylistService, deletions *repository.DeletionRepository) *PlaylistHandler {
	return &PlaylistHandler{service: service, deletions: deletions}
}

func (h *PlaylistHandler) RegisterPublicRoutes(r gin.IRoutes) {
	r.GET("/api/playlists", h.GetAllPublic)
	r.GET("/api/users/:id/playlists", h.GetPublicForUser)
}

func (h *PlaylistHandler) GetPublicForUser(c *gin.Context) {
	userID, err := ParseUserIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	playlists, err := h.service.GetPublicPlaylist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get playlists"})
		return
	}

	c.JSON(http.StatusOK, playlists)
}

func (h *PlaylistHandler) GetAllPublic(c *gin.Context) {
	playlists, err := h.service.GetAllPublic()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get playlists"})
		return
	}
	c.JSON(http.StatusOK, playlists)
}

func (h *PlaylistHandler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.GET("/my/playlists", h.GetMine)
	r.POST("/playlists", h.Create)
	r.POST("/playlists/:id/tracks", h.AddMusic)
	r.DELETE("/playlists/:id", h.Delete)
}

func (h *PlaylistHandler) GetMine(c *gin.Context) {
	playlists, err := h.service.GetByUser(CurrentUserID(c))
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

	playlist, err := h.service.CreateNewPlaylist(req, CurrentUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create playlist"})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

func (h *PlaylistHandler) AddMusic(c *gin.Context) {
	playlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || playlistID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid playlist id"})
		return
	}
	var req struct {
		MusicID int `json:"music_id" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.service.AddMusic(CurrentUserID(c), uint(playlistID), req.MusicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add music to playlist"})
		return
	}
	c.JSON(http.StatusOK, playlist)
}

func (h *PlaylistHandler) Delete(c *gin.Context) {
	playlistID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || playlistID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid playlist id"})
		return
	}
	if err := h.deletions.DeletePlaylist(uint(playlistID), CurrentUserID(c)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete playlist"})
		return
	}
	c.Status(http.StatusNoContent)
}
