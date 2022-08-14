package models

import "github.com/google/uuid"

const UserCredentialTableName = "UserCredentials"

type IdentityType string
type CredentialType string

const (
	IdentityTypeEmail      IdentityType   = "EMAIL"
	CredentialTypePassword CredentialType = "PASSWORD"
)

type UserCredential struct {
	BaseModel
	UserId          *uuid.UUID     `gorm:"column:userId;not null;index" json:"userId"`
	IdentityType    IdentityType   `gorm:"column:identityType;not null;index" json:"identityType"`
	IdentityValue   string         `gorm:"column:identityValue;not null" json:"identityValue"`
	CredentialType  CredentialType `gorm:"column:credentialType;not null;index" json:"credentialType"`
	CredentialValue string         `gorm:"column:credentialValue;not null" json:"credentialValue"`
	IsActive        bool           `gorm:"column:isActive;default:true;index" json:"isActive"`
	User            *User          `gorm:"foreignKey:userId" json:"user,omitempty"`
}

func (m *UserCredential) TableName() string {
	return UserCredentialTableName
}
