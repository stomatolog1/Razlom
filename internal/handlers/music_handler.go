package handlers

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"net/http"
	"github.com/gin-gonic/gin"
)

type MusicHnadler struct {
	service *servise.MusicService
}

func NewMusicHandler(servise *servise.MusicService) *MusicHnadler {
	return &MusicHnadler{service: servise}
}

func (h *MusicHnadler) RegisterRoutes(r *gin.Engine) {
	musics := r.Group("api/music")
	{
		musics.POST("", h.Create)
	}
}

func (h *MusicHnadler) Create(c *gin.Context) {
	var req model.NewMusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	
	user, err := h.service.CreateMusic(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User name already exists"})
	}
	c.JSON(http.StatusCreated, user)
}
