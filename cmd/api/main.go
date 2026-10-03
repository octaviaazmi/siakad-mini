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

	// Mode production → matikan debug Gin
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.Connect(cfg)
	database.Migrate()
	seeders.Run(database.DB)

	r := gin.Default()

	authHandler := handler.NewAuthHandler(database.DB, cfg)
	studentHandler := handler.NewStudentHandler(database.DB)
	courseHandler := handler.NewCourseHandler(database.DB)
	enrollmentHandler := handler.NewEnrollmentHandler(database.DB)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.Login)

		auth := v1.Group("")
		auth.Use(middleware.Auth(cfg))
		{
			auth.GET("/auth/me", authHandler.Me)
			auth.GET("/students/:id", studentHandler.Detail)
			auth.GET("/courses", courseHandler.List)
		}

		admin := v1.Group("")
		admin.Use(middleware.Auth(cfg), middleware.RequireRole("admin"))
		{
			admin.GET("/students", studentHandler.List)
			admin.POST("/students", studentHandler.Create)
			admin.PUT("/students/:id", studentHandler.Update)
			admin.DELETE("/students/:id", studentHandler.Delete)
		}

		mahasiswa := v1.Group("")
		mahasiswa.Use(middleware.Auth(cfg), middleware.RequireRole("mahasiswa"))
		{
			mahasiswa.POST("/enrollments", enrollmentHandler.Create)
			mahasiswa.DELETE("/enrollments/:id", enrollmentHandler.Delete)
		}
	}

	log.Printf("Server jalan di port %s (mode: %s)", cfg.AppPort, cfg.AppEnv)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Server gagal jalan: %v", err)
	}
}
