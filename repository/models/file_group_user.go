package models

const FileGroupUserTableName = "FileGroupUsers"

type FileGroupUser struct {
	BaseModel
	FileGroupId uint64     `gorm:"column:fileGroupId;not null;index" json:"fileGroupId"`
	UserId      uint64     `gorm:"column:userId;not null;index" json:"userId"`
	FileGroup   *FileGroup `gorm:"foreignKey:fileGroupId" json:"fileGroup,omitempty"`
	User        *User      `gorm:"foreignKey:userId" json:"useromitempty"`
}

func (m *FileGroupUser) TableName() string {
	return FileGroupUserTableName
}
