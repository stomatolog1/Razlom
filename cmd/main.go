package main

import (
	"log"
	"os"

	"example/mymodule/internal/handlers"
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
	"example/mymodule/internal/servise"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	sessionSecret := os.Getenv("RAZLOM_SESSION_SECRET")
	sessionManager, err := handlers.NewSessionManager(sessionSecret)
	if err != nil {
		log.Fatalf("configure sessions: %v", err)
	}

	databasePath := os.Getenv("RAZLOM_DB")
	if databasePath == "" {
		databasePath = "razlom.db"
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Music{}, &model.Like{}, &model.Playlist{}, &model.Repost{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	userRepository := repository.NewUserRepository(db)
	userService := servise.NewUserService(userRepository)
	deletionRepository := repository.NewDeletionRepository(db)
	musicRepository := repository.NewMusicRepository(db)
	musicService := servise.NewMusicService(musicRepository)
	likeRepository := repository.NewLikeRepository(db)
	likeService := servise.NewLikeService(likeRepository)
	playlistRepository := repository.NewPlaylistRepository(db)
	playlistService := servise.NewPlaylistService(playlistRepository)
	repostRepository := repository.NewRepostRepository(db)
	repostService := servise.NewRepostService(repostRepository)
	adminService := servise.NewAdminService(userRepository, musicRepository, playlistRepository, deletionRepository)

	uploadDir := os.Getenv("RAZLOM_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		log.Fatalf("prepare upload directory: %v", err)
	}

	router := gin.Default()
	userHandler := handlers.NewUserHandler(userService, sessionManager)
	userHandler.RegisterPublicRoutes(router)
	musicHandler := handlers.NewMusicHandler(musicService, deletionRepository, uploadDir)
	musicHandler.RegisterPublicRoutes(router)
	playlistHandler := handlers.NewPlaylistHandler(playlistService, deletionRepository)
	playlistHandler.RegisterPublicRoutes(router)
	repostHandler := handlers.NewRepostHandler(repostService)
	repostHandler.RegisterPublicRoutes(router)

	protectedAPI := router.Group("/api")
	protectedAPI.Use(sessionManager.RequireAuth())
	userHandler.RegisterProtectedRoutes(protectedAPI)
	musicHandler.RegisterProtectedRoutes(protectedAPI)
	handlers.NewLikeHandler(likeService).RegisterProtectedRoutes(protectedAPI)
	playlistHandler.RegisterProtectedRoutes(protectedAPI)
	repostHandler.RegisterProtectedRoutes(protectedAPI)
	adminAPI := protectedAPI.Group("/admin")
	adminAPI.Use(userHandler.RequireAdmin())
	handlers.NewAdminHandler(adminService, uploadDir).RegisterRoutes(adminAPI)

	router.StaticFile("/", "web/index.html")
	router.StaticFile("/app.js", "web/app.js")
	router.StaticFile("/styles.css", "web/styles.css")
	router.Static("/uploads", uploadDir)

	address := os.Getenv("RAZLOM_ADDR")
	if address == "" {
		address = ":8080"
	}
	if err := router.Run(address); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
