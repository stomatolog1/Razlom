package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxAudioFileSize = 25 << 20

type MusicHandler struct {
	service   *servise.MusicService
	deletions *repository.DeletionRepository
	uploadDir string
}

func NewMusicHandler(service *servise.MusicService, deletions *repository.DeletionRepository, uploadDir string) *MusicHandler {
	return &MusicHandler{service: service, deletions: deletions, uploadDir: uploadDir}
}

func (h *MusicHandler) RegisterPublicRoutes(r gin.IRoutes) {
	r.GET("/api/music", h.GetAll)
	r.GET("/api/users/:id/music", h.GetForUser)
}

func (h *MusicHandler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.POST("/music", h.Create)
	r.DELETE("/music/:id", h.Delete)
}

func (h *MusicHandler) GetAll(c *gin.Context) {
	music, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get music"})
		return
	}

	c.JSON(http.StatusOK, music)
}

func (h *MusicHandler) GetForUser(c *gin.Context) {
	userID, err := ParseUserIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	music, err := h.service.GetAllByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user's music"})
		return
	}
	c.JSON(http.StatusOK, music)
}

func (h *MusicHandler) Create(c *gin.Context) {
	var req model.NewMusicRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAudioFileSize+(1<<20))
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}
	if file.Size == 0 || file.Size > maxAudioFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "audio file must be between 1 byte and 25 MB"})
		return
	}
	extension := strings.ToLower(filepath.Ext(file.Filename))
	switch extension {
	case ".mp3", ".m4a", ".aac", ".ogg", ".wav", ".flac", ".webm":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported audio format"})
		return
	}

	randomName := make([]byte, 16)
	if _, err := rand.Read(randomName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create audio filename"})
		return
	}
	filename := hex.EncodeToString(randomName) + extension
	diskPath := filepath.Join(h.uploadDir, filename)
	if err := os.MkdirAll(h.uploadDir, 0750); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare audio storage"})
		return
	}
	source, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read audio file"})
		return
	}
	defer source.Close()

	destination, err := os.OpenFile(diskPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store audio file"})
		return
	}
	written, copyErr := io.Copy(destination, io.LimitReader(source, maxAudioFileSize+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written > maxAudioFileSize {
		_ = os.Remove(diskPath)
		if copyErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store audio file"})
		} else if closeErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finish storing audio file"})
		} else {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "audio file exceeds 25 MB"})
		}
		return
	}

	audioURL := "/uploads/" + filename
	music, err := h.service.CreateMusic(req, CurrentUserID(c), audioURL)
	if err != nil {
		_ = os.Remove(diskPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create music"})
		return
	}

	c.JSON(http.StatusCreated, music)
}

func (h *MusicHandler) Delete(c *gin.Context) {
	musicID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || musicID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid music id"})
		return
	}

	music, err := h.deletions.DeleteMusic(uint(musicID), CurrentUserID(c))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete music"})
		return
	}

	if err := removeUploadedAudio(h.uploadDir, music.PathToMusic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "track was deleted but its audio file could not be removed"})
		return
	}
	c.Status(http.StatusNoContent)
}

func removeUploadedAudio(uploadDir, audioURL string) error {
	if !strings.HasPrefix(audioURL, "/uploads/") {
		return nil
	}
	filename := filepath.Base(audioURL)
	if filename == "." || filename == string(filepath.Separator) || filename == ".." {
		return errors.New("invalid stored audio path")
	}
	if err := os.Remove(filepath.Join(uploadDir, filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func ParseUserIDParam(c *gin.Context) (uint, error) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		return 0, errors.New("invalid user id")
	}
	return uint(userID), nil
}
