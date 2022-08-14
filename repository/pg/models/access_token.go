package models

import (
	"time"

	"github.com/google/uuid"
)

const AccessTokenTableName = "AccessTokens"

type AccessToken struct {
	BaseModel
	UserId                *uuid.UUID `gorm:"column:userId;not null;index" json:"userId"`
	Token                 string     `gorm:"column:token;not null;index" json:"token"`
	RefreshToken          string     `gorm:"column:refreshToken;not null;index" json:"refreshToken"`
	TokenExpiredAt        time.Time  `gorm:"column:tokenExpiredAt;not null" json:"tokenExpiredAt"`
	RefreshTokenExpiredAt time.Time  `gorm:"column:refreshTokenExpiredAt;not null" json:"refreshTokenExpiredAt"`
	User                  *User      `gorm:"foreignKey:userId" json:"user,omitempty"`
}

func (m *AccessToken) TableName() string {
	return AccessTokenTableName
}
