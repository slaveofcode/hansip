package auth

import (
	"errors"
	"net/http"

	"filippo.io/age"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/repository/pg/models"
	"github.com/slaveofcode/securi/utils/aes"
	"gorm.io/gorm"
)

type RegisterParam struct {
	Name            string `json:"name" binding:"required"`
	Alias           string `json:"alias" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,eqfield=ConfirmPassword"`
	ConfirmPassword string `json:"cpassword" binding:"required"`
}

func Register(repo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		var bodyParams RegisterParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		if bodyParams.Password != bodyParams.ConfirmPassword {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Password mismatch",
			})
			return
		}

		// create user
		db := repo.GetDB()

		err := db.Transaction(func(tx *gorm.DB) error {
			var findExtCred models.UserCredential
			res := tx.Where(&models.UserCredential{
				IdentityType:  models.IdentityTypeEmail,
				IdentityValue: bodyParams.Email,
			}).First(&findExtCred)

			if res.RowsAffected > 0 {
				return errors.New("existing email already exist")
			}

			user := models.User{
				Name:  bodyParams.Name,
				Alias: bodyParams.Alias,
			}
			res = tx.Create(&user)
			if res.Error != nil {
				return res.Error
			}

			passEnc, err := bcrypt.GenerateFromPassword([]byte(bodyParams.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			userCred := models.UserCredential{
				UserId:          &user.ID,
				IdentityType:    models.IdentityTypeEmail,
				IdentityValue:   bodyParams.Email,
				CredentialType:  models.CredentialTypePassword,
				CredentialValue: string(passEnc),
			}

			res = tx.Create(&userCred)
			if res.Error != nil {
				return res.Error
			}

			identity, err := age.GenerateX25519Identity()
			if err != nil {
				return err
			}

			publicKey := identity.Recipient().String()
			encPrivateKey := aes.Encrypt(bodyParams.Password, identity.String())
			userKey := models.UserKey{
				UserId:  &user.ID,
				Public:  publicKey,
				Private: encPrivateKey,
			}

			res = tx.Create(&userKey)
			if res.Error != nil {
				return res.Error
			}

			return nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed registering account: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  true,
			"message": "Registered",
		})
	}
}
