package handlers

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/servise"
	"net/http"
	"github.com/gin-gonic/gin"

)

type UserHandler struct {
	servise *servise.UserService
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine){
	users := r.Group("api/users")
	{
		users.POST("", h.Create)
	}
}

func NewUserHandler(service *servise.UserService) *UserHandler{
	return &UserHandler{servise: service}
}

func (h *UserHandler) Create(c *gin.Context){
	var req model.CreateUserRequest
	if err:= c.ShouldBindJSON(&req); err != nil{
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	user, err := h.servise.CreateUser(req)
	if err!=nil{
		c.JSON(http.StatusConflict, gin.H{"error": "User name already exists"})
	}
	c.JSON(http.StatusCreated, user)
}