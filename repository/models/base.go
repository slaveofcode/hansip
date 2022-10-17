package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;not null" json:"id"`
	CreatedAt time.Time      `gorm:"column:createdAt;index;<-:create" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updatedAt;index" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deletedAt;index" json:"deletedAt"`
}
