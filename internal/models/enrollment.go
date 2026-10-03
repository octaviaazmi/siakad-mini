package models

import "time"

type Enrollment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentID     uint      `gorm:"not null;uniqueIndex:idx_student_course_year" json:"student_id"`
	CourseID      uint      `gorm:"not null;uniqueIndex:idx_student_course_year" json:"course_id"`
	TahunAkademik string    `gorm:"size:20;not null;uniqueIndex:idx_student_course_year" json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`

	Student Student `gorm:"foreignKey:StudentID" json:"-"`
	Course  Course  `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}
