package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"siakad-mini/internal/models"
	"siakad-mini/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EnrollmentHandler struct {
	DB *gorm.DB
}

func NewEnrollmentHandler(db *gorm.DB) *EnrollmentHandler {
	return &EnrollmentHandler{DB: db}
}

var reTahunAkademik = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

type createEnrollmentRequest struct {
	CourseID      uint   `json:"course_id" binding:"required"`
	TahunAkademik string `json:"tahun_akademik" binding:"required"`
}

// =========================================================
// POST /api/v1/enrollments
// =========================================================
func (h *EnrollmentHandler) Create(c *gin.Context) {
	uidAny, _ := c.Get("user_id")
	uid, _ := uidAny.(uint)

	var student models.Student
	if err := h.DB.Where("user_id = ?", uid).First(&student).Error; err != nil {
		utils.JSONError(c, http.StatusForbidden, "Data mahasiswa tidak ditemukan", nil)
		return
	}

	var req createEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONError(c, http.StatusUnprocessableEntity, "Validasi gagal", map[string]string{
			"body": err.Error(),
		})
		return
	}

	if !reTahunAkademik.MatchString(req.TahunAkademik) {
		utils.JSONError(c, http.StatusUnprocessableEntity, "Validasi gagal", map[string]string{
			"tahun_akademik": "Format harus YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap",
		})
		return
	}

	var course models.Course
	if err := h.DB.First(&course, req.CourseID).Error; err != nil {
		utils.JSONError(c, http.StatusUnprocessableEntity, "Validasi gagal", map[string]string{
			"course_id": "Mata kuliah tidak ditemukan",
		})
		return
	}

	var dupCount int64
	h.DB.Model(&models.Enrollment{}).
		Where("student_id = ? AND course_id = ? AND tahun_akademik = ?",
			student.ID, req.CourseID, req.TahunAkademik).
		Count(&dupCount)
	if dupCount > 0 {
		utils.JSONError(c, http.StatusConflict,
			"Mata kuliah sudah pernah diambil pada tahun akademik ini", nil)
		return
	}

	batas := hitungBatasSKS(student.IPKTerakhir)

	var sumResult struct {
		Total int64
	}
	h.DB.Table("enrollments").
		Select("COALESCE(SUM(courses.sks), 0) AS total").
		Joins("JOIN courses ON courses.id = enrollments.course_id").
		Where("enrollments.student_id = ? AND enrollments.tahun_akademik = ?",
			student.ID, req.TahunAkademik).
		Scan(&sumResult)

	if int(sumResult.Total)+course.SKS > batas {
		sisa := batas - int(sumResult.Total)
		utils.JSONError(c, http.StatusUnprocessableEntity,
			fmt.Sprintf("Total SKS melebihi batas. Sisa SKS Anda: %d", sisa), nil)
		return
	}

	var enrollment models.Enrollment
	txErr := h.DB.Transaction(func(tx *gorm.DB) error {
		var lockedCourse models.Course
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&lockedCourse, req.CourseID).Error; err != nil {
			return err
		}

		var terisi int64
		tx.Model(&models.Enrollment{}).
			Where("course_id = ?", lockedCourse.ID).
			Count(&terisi)

		if int(terisi) >= lockedCourse.Kuota {
			return errors.New("KUOTA_PENUH")
		}

		enrollment = models.Enrollment{
			StudentID:     student.ID,
			CourseID:      lockedCourse.ID,
			TahunAkademik: req.TahunAkademik,
		}
		return tx.Create(&enrollment).Error
	})

	if txErr != nil {
		if txErr.Error() == "KUOTA_PENUH" {
			utils.JSONError(c, http.StatusUnprocessableEntity,
				"Kuota mata kuliah sudah penuh", nil)
			return
		}
		utils.JSONError(c, http.StatusInternalServerError,
			"Gagal menyimpan enrollment", nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Berhasil mengambil mata kuliah", gin.H{
		"id":             enrollment.ID,
		"student_id":     enrollment.StudentID,
		"course_id":      enrollment.CourseID,
		"kode_mk":        course.KodeMK,
		"nama_mk":        course.NamaMK,
		"sks":            course.SKS,
		"tahun_akademik": enrollment.TahunAkademik,
	})
}

// =========================================================
// DELETE /api/v1/enrollments/{id}
// =========================================================
func (h *EnrollmentHandler) Delete(c *gin.Context) {
	uidAny, _ := c.Get("user_id")
	uid, _ := uidAny.(uint)

	// Cari student berdasarkan user login
	var student models.Student
	if err := h.DB.Where("user_id = ?", uid).First(&student).Error; err != nil {
		utils.JSONError(c, http.StatusForbidden, "Data mahasiswa tidak ditemukan", nil)
		return
	}

	// Parse ID
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id < 1 {
		utils.JSONError(c, http.StatusNotFound, "Enrollment tidak ditemukan", nil)
		return
	}

	// Cari enrollment
	var enrollment models.Enrollment
	if err := h.DB.First(&enrollment, id).Error; err != nil {
		utils.JSONError(c, http.StatusNotFound, "Enrollment tidak ditemukan", nil)
		return
	}

	// Cek ownership
	if enrollment.StudentID != student.ID {
		utils.JSONError(c, http.StatusForbidden, "Akses ditolak", nil)
		return
	}

	// Hapus
	if err := h.DB.Delete(&enrollment).Error; err != nil {
		utils.JSONError(c, http.StatusInternalServerError, "Gagal menghapus enrollment", nil)
		return
	}

	c.Status(http.StatusNoContent)
}
