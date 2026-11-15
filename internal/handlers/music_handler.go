package handlers

import (
	"net/http"

	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
)

type MusicHandler struct {
	service *servise.MusicService
}

func NewMusicHandler(service *servise.MusicService) *MusicHandler {
	return &MusicHandler{service: service}
}

func (h *MusicHandler) RegisterRoutes(r *gin.Engine) {
	music := r.Group("/api/music")
	music.GET("", h.GetAll)
	music.POST("", h.Create)
}

func (h *MusicHandler) GetAll(c *gin.Context) {
	music, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get music"})
		return
	}

	c.JSON(http.StatusOK, music)
}

func (h *MusicHandler) Create(c *gin.Context) {
	var req model.NewMusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	music, err := h.service.CreateMusic(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create music"})
		return
	}

	c.JSON(http.StatusCreated, music)
}
