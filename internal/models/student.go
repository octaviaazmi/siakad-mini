package models

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	NIM         string         `gorm:"uniqueIndex;size:12;not null" json:"nim"`
	Nama        string         `gorm:"size:100;not null" json:"nama"`
	Prodi       string         `gorm:"size:100;not null" json:"prodi"`
	Angkatan    int            `gorm:"not null" json:"angkatan"`
	IPKTerakhir float64        `gorm:"column:ipk_terakhir;type:numeric(3,2);default:0" json:"ipk_terakhir"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	User        User         `gorm:"foreignKey:UserID" json:"-"`
	Enrollments []Enrollment `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
}
