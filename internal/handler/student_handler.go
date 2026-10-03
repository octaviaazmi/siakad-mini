package handler

import (
	"net/http"
	"strconv"

	"siakad-mini/internal/models"
	"siakad-mini/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StudentHandler struct {
	DB *gorm.DB
}

func NewStudentHandler(db *gorm.DB) *StudentHandler {
	return &StudentHandler{DB: db}
}

// GET /api/v1/students
func (h *StudentHandler) List(c *gin.Context) {
	// --- Pagination ---
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

	// --- Filters ---
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

	// --- Sorting ---
	switch sort {
	case "nama":
		q = q.Order("nama ASC")
	case "-ipk_terakhir":
		q = q.Order("ipk_terakhir DESC")
	default:
		q = q.Order("id ASC")
	}

	// --- Count total ---
	var total int64
	if err := q.Count(&total).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal menghitung data", nil)
		return
	}

	// --- Ambil data ---
	offset := (page - 1) * perPage
	var students []models.Student
	if err := q.Limit(perPage).Offset(offset).Find(&students).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal mengambil data", nil)
		return
	}

	// --- Format response ---
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
