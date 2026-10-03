package main

import (
	"log"

	"siakad-mini/internal/config"
	"siakad-mini/internal/database"
	"siakad-mini/internal/handler"
	"siakad-mini/internal/middleware"
	"siakad-mini/internal/seeders"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database.Connect(cfg)
	database.Migrate()
	seeders.Run(database.DB)

	r := gin.Default()

	authHandler := handler.NewAuthHandler(database.DB, cfg)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.Login)

		auth := v1.Group("")
		auth.Use(middleware.Auth(cfg))
		{
			auth.GET("/auth/me", authHandler.Me)
		}
	}

	log.Printf("Server jalan di port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server gagal jalan: %v", err)
	}
}
