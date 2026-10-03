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
	studentHandler := handler.NewStudentHandler(database.DB)

	v1 := r.Group("/api/v1")
	{
		// publik
		v1.POST("/auth/login", authHandler.Login)

		// butuh token
		auth := v1.Group("")
		auth.Use(middleware.Auth(cfg))
		{
			auth.GET("/auth/me", authHandler.Me)
		}

		// admin only
		admin := v1.Group("")
		admin.Use(middleware.Auth(cfg), middleware.RequireRole("admin"))
		{
			admin.GET("/students", studentHandler.List)
			admin.POST("/students", studentHandler.Create)
		}
	}

	log.Printf("Server jalan di port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server gagal jalan: %v", err)
	}
}
