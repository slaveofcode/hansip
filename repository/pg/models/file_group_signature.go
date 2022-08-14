package models

import (
	"github.com/google/uuid"
)

const FileGroupSignatureTableName = "FileGroupSignatures"

type FileGroupSignature struct {
	BaseModel
	FileGroupId *uuid.UUID `gorm:"column:fileGroupId;not null;index" json:"fileGroupId"`
	UserKeyId   *uuid.UUID `gorm:"column:userKeyId;not null;index" json:"userKeyId"`
	FileGroup   *FileGroup `gorm:"foreignKey:fileGroupId" json:"fileGroup,omitempty"`
	UserKey     *UserKey   `gorm:"foreignKey:userKeyId" json:"userKey,omitempty"`
}

func (m *FileGroupSignature) TableName() string {
	return FileGroupSignatureTableName
}
