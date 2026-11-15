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

type UserHandler struct {
	service  *servise.UserService
	sessions *SessionManager
}

func NewUserHandler(service *servise.UserService, sessions *SessionManager) *UserHandler {
	return &UserHandler{service: service, sessions: sessions}
}

func (h *UserHandler) RegisterPublicRoutes(r gin.IRoutes) {
	r.POST("/api/users", h.Create)
	r.POST("/api/sessions", h.Login)
	r.GET("/api/users/:id", h.GetPublicProfile)
}

func (h *UserHandler) RegisterProtectedRoutes(r gin.IRoutes) {
	r.GET("/session", h.GetSession)
	r.DELETE("/session", h.Logout)
}

func (h *UserHandler) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := h.service.GetByID(CurrentUserID(c))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				h.sessions.ClearCookie(c)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user no longer exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify administrator"})
			}
			c.Abort()
			return
		}
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "administrator access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.CreateUser(req)
	if err != nil {
		if errors.Is(err, servise.ErrLoginTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "login is already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	h.sessions.SetCookie(c, user.ID)
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Authenticate(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, servise.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}
	h.sessions.SetCookie(c, user.ID)
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetSession(c *gin.Context) {
	user, err := h.service.GetByID(CurrentUserID(c))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.sessions.ClearCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user no longer exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load session"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Logout(c *gin.Context) {
	h.sessions.ClearCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) GetPublicProfile(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.service.GetByID(uint(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	c.JSON(http.StatusOK, user)
}
