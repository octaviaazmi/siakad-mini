package handler

import (
	"net/http"
	"sync"
	"time"

	"siakad-mini/internal/config"
	"siakad-mini/internal/models"
	"siakad-mini/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// ---- Simple in-memory rate limiter (5 failed / minute / IP) ----
var (
	loginAttempts = make(map[string][]time.Time)
	loginMutex    sync.Mutex
)

func isRateLimited(ip string) bool {
	loginMutex.Lock()
	defer loginMutex.Unlock()

	cutoff := time.Now().Add(-time.Minute)
	var recent []time.Time
	for _, t := range loginAttempts[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	loginAttempts[ip] = recent
	return len(recent) >= 5
}

func recordFailure(ip string) {
	loginMutex.Lock()
	defer loginMutex.Unlock()
	loginAttempts[ip] = append(loginAttempts[ip], time.Now())
}

func (h *AuthHandler) Login(c *gin.Context) {
	ip := c.ClientIP()

	if isRateLimited(ip) {
		utils.JSONError(c, http.StatusTooManyRequests, "Terlalu banyak percobaan login. Coba lagi nanti.", nil)
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusUnprocessableEntity, "Validasi gagal", map[string]string{
			"body": err.Error(),
		})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		recordFailure(ip)
		utils.JSONError(c, http.StatusUnauthorized, "Email atau password salah", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		recordFailure(ip)
		utils.JSONError(c, http.StatusUnauthorized, "Email atau password salah", nil)
		return
	}

	// pastikan mahasiswa yang sudah soft-delete tidak bisa login
	if user.Role == "mahasiswa" {
		var s models.Student
		if err := h.DB.Where("user_id = ?", user.ID).First(&s).Error; err != nil {
			recordFailure(ip)
			utils.JSONError(c, http.StatusUnauthorized, "Akun tidak aktif", nil)
			return
		}
	}

	token, expiresIn, err := utils.GenerateToken(h.Cfg.JWTSecret, user.ID, user.Email, user.Role, h.Cfg.JWTExpiredHours)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal membuat token", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Login berhasil", gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	uidAny, _ := c.Get("user_id")
	uid, _ := uidAny.(uint)

	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		utils.JSONError(c, http.StatusUnauthorized, "User tidak ditemukan", nil)
		return
	}

	resp := gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "mahasiswa" {
		var s models.Student
		if err := h.DB.Where("user_id = ?", user.ID).First(&s).Error; err == nil {
			resp["student"] = gin.H{
				"nim":      s.NIM,
				"nama":     s.Nama,
				"prodi":    s.Prodi,
				"angkatan": s.Angkatan,
			}
		}
	}

	utils.JSONSuccess(c, http.StatusOK, "Profil berhasil diambil", resp)
}
