package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"siakad-mini/internal/models"
	"siakad-mini/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type StudentHandler struct {
	DB *gorm.DB
}

func NewStudentHandler(db *gorm.DB) *StudentHandler {
	return &StudentHandler{DB: db}
}

// ---- Validasi helpers ----

var (
	reNIM   = regexp.MustCompile(`^\d{12}$`)
	reEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// =========================================================
// GET /api/v1/students
// =========================================================
func (h *StudentHandler) List(c *gin.Context) {
	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	// Filters
	prodi := c.Query("prodi")
	angkatan := c.Query("angkatan")
	search := c.Query("search")
	sort := c.Query("sort")

	q := h.DB.Model(&models.Student{})

	if prodi != "" {
		q = q.Where("prodi = ?", prodi)
	}
	if angkatan != "" {
		q = q.Where("angkatan = ?", angkatan)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("nim ILIKE ? OR nama ILIKE ?", like, like)
	}

	switch sort {
	case "nama":
		q = q.Order("nama ASC")
	case "-ipk_terakhir":
		q = q.Order("ipk_terakhir DESC")
	default:
		q = q.Order("id ASC")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal menghitung data", nil)
		return
	}

	offset := (page - 1) * perPage
	var students []models.Student
	if err := q.Limit(perPage).Offset(offset).Find(&students).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal mengambil data", nil)
		return
	}

	type studentItem struct {
		ID          uint    `json:"id"`
		NIM         string  `json:"nim"`
		Nama        string  `json:"nama"`
		Prodi       string  `json:"prodi"`
		Angkatan    int     `json:"angkatan"`
		IPKTerakhir float64 `json:"ipk_terakhir"`
	}

	items := make([]studentItem, 0, len(students))
	for _, s := range students {
		items = append(items, studentItem{
			ID:          s.ID,
			NIM:         s.NIM,
			Nama:        s.Nama,
			Prodi:       s.Prodi,
			Angkatan:    s.Angkatan,
			IPKTerakhir: s.IPKTerakhir,
		})
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))

	meta := gin.H{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}

	utils.JSONSuccessWithMeta(c, http.StatusOK, "Data mahasiswa berhasil diambil", items, meta)
}

// =========================================================
// POST /api/v1/students
// =========================================================
type createStudentRequest struct {
	NIM         string   `json:"nim" binding:"required"`
	Nama        string   `json:"nama" binding:"required"`
	Email       string   `json:"email" binding:"required"`
	Prodi       string   `json:"prodi" binding:"required"`
	Angkatan    int      `json:"angkatan" binding:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}

func (h *StudentHandler) Create(c *gin.Context) {
	var req createStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusUnprocessableEntity, "Validasi gagal", map[string]string{
			"body": err.Error(),
		})
		return
	}

	errors := map[string]string{}

	if !reNIM.MatchString(req.NIM) {
		errors["nim"] = "NIM harus 12 digit angka"
	}
	if !reEmail.MatchString(req.Email) {
		errors["email"] = "Format email tidak valid"
	}

	currentYear := time.Now().Year()
	if req.Angkatan < 1900 || req.Angkatan > currentYear {
		errors["angkatan"] = "Angkatan harus 4 digit dan tidak melebihi tahun berjalan"
	}

	ipk := 0.0
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
		if ipk < 0 || ipk > 4 {
			errors["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
		}
	}

	var count int64
	h.DB.Model(&models.Student{}).Where("nim = ?", req.NIM).Count(&count)
	if count > 0 {
		errors["nim"] = "NIM sudah terdaftar"
	}

	h.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		errors["email"] = "Email sudah terdaftar"
	}

	if len(errors) > 0 {
		utils.JSONError(c, http.StatusUnprocessableEntity, "Validasi gagal", errors)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NIM), bcrypt.DefaultCost)
	if err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal hash password", nil)
		return
	}

	var student models.Student

	txErr := h.DB.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			Email:    req.Email,
			Password: string(hash),
			Role:     "mahasiswa",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		student = models.Student{
			UserID:      user.ID,
			NIM:         req.NIM,
			Nama:        req.Nama,
			Prodi:       req.Prodi,
			Angkatan:    req.Angkatan,
			IPKTerakhir: ipk,
		}
		if err := tx.Create(&student).Error; err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal menyimpan data", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Mahasiswa berhasil ditambahkan", gin.H{
		"id":           student.ID,
		"nim":          student.NIM,
		"nama":         student.Nama,
		"prodi":        student.Prodi,
		"angkatan":     student.Angkatan,
		"ipk_terakhir": student.IPKTerakhir,
		"email":        req.Email,
	})
}
