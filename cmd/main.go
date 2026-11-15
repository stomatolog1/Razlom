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
	databasePath := os.Getenv("RAZLOM_DB")
	if databasePath == "" {
		databasePath = "razlom.db"
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.Music{}, &model.Like{}, &model.Playlist{}, &model.Repost{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	musicRepository := repository.NewMusicRepository(db)
	musicService := servise.NewMusicService(musicRepository)
	likeRepository := repository.NewLikeRepository(db)
	likeService := servise.NewLikeService(likeRepository)
	playlistRepository := repository.NewPlaylistRepository(db)
	playlistService := servise.NewPlaylistService(playlistRepository)
	repostRepository := repository.NewRepostRepository(db)
	repostService := servise.NewRepostService(repostRepository)
	router := gin.Default()
	handlers.NewMusicHandler(musicService).RegisterRoutes(router)
	handlers.NewLikeHandler(likeService).RegisterRoutes(router)
	handlers.NewPlaylistHandler(playlistService).RegisterRoutes(router)
	handlers.NewRepostHandler(repostService).RegisterRoutes(router)

	router.StaticFile("/", "web/index.html")
	router.StaticFile("/app.js", "web/app.js")
	router.StaticFile("/styles.css", "web/styles.css")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
