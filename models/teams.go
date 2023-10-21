package models

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	ID        int64  `gorm:"column:id"`
	Name      string `gorm:"column:name;type:varchar(100)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
