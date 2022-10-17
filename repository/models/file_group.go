package models

import (
	"time"
)

type ArchiveType string

const (
	ArchiveTypeZIP ArchiveType = "ZIP"
	ArchiveTypeRAR ArchiveType = "RAR"
	ArchiveTypeTAR ArchiveType = "TAR"
)

const FileGroupTableName = "FileGroups"

type FileGroup struct {
	BaseModel
	UserId                uint64      `gorm:"column:userId;not null;index" json:"userId"`
	TotalFiles            int         `gorm:"column:totalFiles;index" json:"totalFiles"`
	ArchiveType           ArchiveType `gorm:"column:archiveType;index" json:"archiveType"`
	ArchivePasscode       string      `gorm:"column:archivePasscode" json:"archivePasscode"`
	MaxDownload           int         `gorm:"column:maxDownload" json:"maxDownload"`
	DeleteAtDownloadTimes int         `gorm:"column:deleteAtDownloadTimes;index" json:"deleteAtDownloadTimes"`
	ExpiredAt             *time.Time  `gorm:"column:expiredAt;index" json:"expiredAt"`
	BundledAt             *time.Time  `gorm:"column:bundledAt;index" json:"bundledAt"`
	User                  *User       `gorm:"foreignKey:userId" json:"user,omitempty"`

	FileKey string `gorm:"column:fileKey" json:"fileKey"`

	FileItems []FileItem
}

func (m *FileGroup) TableName() string {
	return FileGroupTableName
}
