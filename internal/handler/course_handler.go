package handler

import (
	"net/http"
	"strconv"

	"siakad-mini/internal/models"
	"siakad-mini/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CourseHandler struct {
	DB *gorm.DB
}

func NewCourseHandler(db *gorm.DB) *CourseHandler {
	return &CourseHandler{DB: db}
}

// =========================================================
// GET /api/v1/courses
// =========================================================
func (h *CourseHandler) List(c *gin.Context) {
	semester := c.Query("semester")
	search := c.Query("search")
	available := c.Query("available")

	q := h.DB.Model(&models.Course{})

	if semester != "" {
		q = q.Where("semester = ?", semester)
	}

	if search != "" {
		like := "%" + search + "%"
		q = q.Where("kode_mk ILIKE ? OR nama_mk ILIKE ?", like, like)
	}

	var courses []models.Course
	if err := q.Order("id ASC").Find(&courses).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal mengambil data", nil)
		return
	}

	type courseItem struct {
		ID        uint   `json:"id"`
		KodeMK    string `json:"kode_mk"`
		NamaMK    string `json:"nama_mk"`
		SKS       int    `json:"sks"`
		Semester  int    `json:"semester"`
		Kuota     int    `json:"kuota"`
		Terisi    int    `json:"terisi"`
		SisaKuota int    `json:"sisa_kuota"`
	}

	items := make([]courseItem, 0, len(courses))

	for _, cr := range courses {
		var terisi int64
		h.DB.Model(&models.Enrollment{}).
			Where("course_id = ?", cr.ID).
			Count(&terisi)

		sisa := cr.Kuota - int(terisi)

		// filter available
		if available == "true" && sisa <= 0 {
			continue
		}

		items = append(items, courseItem{
			ID:        cr.ID,
			KodeMK:    cr.KodeMK,
			NamaMK:    cr.NamaMK,
			SKS:       cr.SKS,
			Semester:  cr.Semester,
			Kuota:     cr.Kuota,
			Terisi:    int(terisi),
			SisaKuota: sisa,
		})
	}

	utils.JSONSuccess(c, http.StatusOK, "Data mata kuliah berhasil diambil", items)
}

var _ = strconv.Atoi
