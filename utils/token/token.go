package token

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/slaveofcode/hansip/repository/models"
	"gorm.io/gorm"
)

const (
	ACCESS_TOKEN_LENGTH  = 64
	REFRESH_TOKEN_LENGTH = 32
	MAX_GEN_TOKEN_ITER   = 1_000_000
)

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

type TokenInfo struct {
	AccessToken  string
	RefreshToken string
}

func GetFreshTokens(db *gorm.DB) (*TokenInfo, error) {

	for i := 0; i < MAX_GEN_TOKEN_ITER; i++ {
		newAccToken := GenerateSecureToken(ACCESS_TOKEN_LENGTH)
		newRefToken := GenerateSecureToken(REFRESH_TOKEN_LENGTH)

		var acct models.AccessToken
		res := db.Where(`token = ? OR "refreshToken" = ?`, newAccToken, newRefToken).First(&acct)

		if res.RowsAffected == 0 {
			return &TokenInfo{
				AccessToken:  newAccToken,
				RefreshToken: newRefToken,
			}, nil
		}
	}

	return nil, errors.New("unable to generate unique access token")
}
