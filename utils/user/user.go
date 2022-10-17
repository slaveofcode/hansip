package user

import (
	"fmt"
	"strings"
	"time"

	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
)

func GetUserFromHeaderAuth(pgRepo repository.Repository, token string) (*models.User, error) {
	bearers := strings.Split(token, " ")
	bearer := bearers[1]

	db := pgRepo.GetDB()

	var acct models.AccessToken
	res := db.Preload("User").Where(&models.AccessToken{
		Token: bearer,
	}).First(&acct)

	if res.RowsAffected <= 0 {
		return nil, fmt.Errorf("user not found")
	}

	if acct.TokenExpiredAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return acct.User, nil
}
