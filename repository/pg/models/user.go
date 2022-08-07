package models

import "gorm.io/datatypes"

const UserTableName = "Users"

type User struct {
	BaseModel
	Name     string         `gorm:"column:name;index" json:"name"`
	Metadata datatypes.JSON `gorm:"column:metadata;index" json:"metadata"`
}

func (m *User) TableName() string {
	return UserTableName
}
