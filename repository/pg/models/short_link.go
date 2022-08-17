package models

import (
	"github.com/google/uuid"
)

const ShortLinkTableName = "ShortLinks"

type ShortLink struct {
	BaseModel
	FileGroupId *uuid.UUID `gorm:"column:fileGroupId;not null;index" json:"fileGroupId"`
	ShortCode   string     `gorm:"column:shortCode;unique;not null;index" json:"shortCode"`
	PIN         string     `gorm:"column:pin" json:"pin"`
	FileGroup   *FileGroup `gorm:"foreignKey:fileGroupId" json:"fileGroup,omitempty"`
}

func (m *ShortLink) TableName() string {
	return ShortLinkTableName
}
