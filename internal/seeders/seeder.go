package seeders

import (
	"fmt"
	"log"
	"math/rand"

	"siakad-mini/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	// 1 admin
	seedAdmin(db)

	// 20 mahasiswa
	seedMahasiswa(db, 20)

	// 10 mata kuliah
	seedCourses(db, 10)

	log.Println("Seeding selesai")
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin12345"), bcrypt.DefaultCost)

	admin := models.User{
		Email:    "admin@siakad.test",
		Password: string(hash),
		Role:     "admin",
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Println("Gagal seed admin:", err)
		return
	}
	log.Println("Admin dibuat: admin@siakad.test / admin12345")
}

func seedMahasiswa(db *gorm.DB, n int) {
	prodiList := []string{"Sistem Informasi", "Teknik Informatika", "Manajemen Informatika"}

	for i := 1; i <= n; i++ {
		nim := fmt.Sprintf("187221%06d", i)
		email := fmt.Sprintf("mhs%d@siakad.test", i)

		var existing models.User
		if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
			continue
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(nim), bcrypt.DefaultCost)

		user := models.User{
			Email:    email,
			Password: string(hash),
			Role:     "mahasiswa",
		}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Gagal seed user mhs%d: %v", i, err)
			continue
		}

		student := models.Student{
			UserID:      user.ID,
			NIM:         nim,
			Nama:        fmt.Sprintf("Mahasiswa %d", i),
			Prodi:       prodiList[i%len(prodiList)],
			Angkatan:    2022 + (i % 3),
			IPKTerakhir: 2.00 + rand.Float64()*2.00, // 2.00–4.00
		}
		// bulatkan 2 desimal
		student.IPKTerakhir = float64(int(student.IPKTerakhir*100)) / 100

		if err := db.Create(&student).Error; err != nil {
			log.Printf("Gagal seed student mhs%d: %v", i, err)
		}
	}
	log.Printf("%d mahasiswa di-seed", n)
}

func seedCourses(db *gorm.DB, n int) {
	for i := 1; i <= n; i++ {
		kode := fmt.Sprintf("MK%03d", i)
		var existing models.Course
		if err := db.Where("kode_mk = ?", kode).First(&existing).Error; err == nil {
			continue
		}

		course := models.Course{
			KodeMK:   kode,
			NamaMK:   fmt.Sprintf("Mata Kuliah %d", i),
			SKS:      2 + (i % 3), // 2–4 SKS
			Semester: 1 + (i % 8),
			Kuota:    30,
		}
		if err := db.Create(&course).Error; err != nil {
			log.Printf("Gagal seed course %s: %v", kode, err)
		}
	}
	log.Printf("%d mata kuliah di-seed", n)
}
