package models

type PreviewAsType string

const (
	PreviewAsImage    PreviewAsType = "IMAGE"
	PreviewAsDocument PreviewAsType = "DOCUMENT"
	PreviewAsAudio    PreviewAsType = "AUDIO"
	PreviewAsVideo    PreviewAsType = "VIDEO"
	PreviewAsArchive  PreviewAsType = "ARCHIVE"
	PreviewAsFont     PreviewAsType = "FONT"
	PreviewAsBinary   PreviewAsType = "BINARY"
)

const FileItemTableName = "FileItems"

type FileItem struct {
	BaseModel
	FileGroupId uint64        `gorm:"column:fileGroupId;not null;index" json:"fileGroupId"`
	Filename    string        `gorm:"column:fileName;not null;index" json:"fileName"`
	Realname    string        `gorm:"column:realName;not null;index" json:"realName"`
	PreviewAs   PreviewAsType `gorm:"column:previewAs;not null;index" json:"previewAs"`
	SizeInBytes int64         `gorm:"column:sizeInBytes" json:"sizeInBytes"`
	FileGroup   *FileGroup    `gorm:"foreignKey:fileGroupId" json:"fileGroup,omitempty"`
}

func (m *FileItem) TableName() string {
	return FileItemTableName
}
