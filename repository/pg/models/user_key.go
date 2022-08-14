package models

import (
	"github.com/google/uuid"
)

const UserKeyTableName = "UserKeys"

type UserKey struct {
	BaseModel
	UserId  *uuid.UUID `gorm:"column:userId;not null;index" json:"userId"`
	Public  string     `gorm:"column:public;not null;index" json:"public"`
	Private string     `gorm:"column:private;not null;index" json:"private"`
	User    *User      `gorm:"foreignKey:userId" json:"user,omitempty"`
}

func (m *UserKey) TableName() string {
	return UserKeyTableName
}
