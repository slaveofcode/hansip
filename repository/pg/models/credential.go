package models

import "gorm.io/datatypes"

const CredentialTableName = "Credentials"

type Credential struct {
	BaseModel
	Name     string         `gorm:"column:name;index" json:"name"`
	Metadata datatypes.JSON `gorm:"column:metadata;index" json:"metadata"`
}

func (m *Credential) TableName() string {
	return CredentialTableName
}
