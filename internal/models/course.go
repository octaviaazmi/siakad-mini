package models

import "time"

type Course struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	KodeMK    string    `gorm:"uniqueIndex;size:20;not null" json:"kode_mk"`
	NamaMK    string    `gorm:"size:150;not null" json:"nama_mk"`
	SKS       int       `gorm:"not null" json:"sks"`
	Semester  int       `gorm:"not null" json:"semester"`
	Kuota     int       `gorm:"not null" json:"kuota"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Enrollments []Enrollment `gorm:"foreignKey:CourseID" json:"-"`
}
