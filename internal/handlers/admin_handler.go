package handlers

import (
	"errors"
	"net/http"

	"example/mymodule/internal/repository"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	service   *servise.AdminService
	uploadDir string
}

func NewAdminHandler(service *servise.AdminService, uploadDir string) *AdminHandler {
	return &AdminHandler{service: service, uploadDir: uploadDir}
}

func (h *AdminHandler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/users", h.GetUsers)
	r.GET("/music", h.GetMusic)
	r.GET("/playlists", h.GetPlaylists)
	r.DELETE("/users/:id", h.DeleteUser)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) GetMusic(c *gin.Context) {
	music, err := h.service.GetMusic()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get music"})
		return
	}
	c.JSON(http.StatusOK, music)
}

func (h *AdminHandler) GetPlaylists(c *gin.Context) {
	playlists, err := h.service.GetPlaylists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get playlists"})
		return
	}
	c.JSON(http.StatusOK, playlists)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := ParseUserIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uploadPaths, err := h.service.DeleteUser(userID, CurrentUserID(c))
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, repository.ErrAdminRequired):
			c.JSON(http.StatusForbidden, gin.H{"error": "administrator access required"})
		case errors.Is(err, repository.ErrSelfDeletion):
			c.JSON(http.StatusConflict, gin.H{"error": "cannot delete your current administrator account"})
		case errors.Is(err, repository.ErrLastAdminDeletion):
			c.JSON(http.StatusConflict, gin.H{"error": "cannot delete the last administrator"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		}
		return
	}
	for _, audioURL := range uploadPaths {
		if err := removeUploadedAudio(h.uploadDir, audioURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user was deleted but an audio file could not be removed"})
			return
		}
	}
	c.Status(http.StatusNoContent)
}
