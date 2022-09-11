package download

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/age_encryption"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/repository/pg/models"
	"github.com/slaveofcode/hansip/utils/aes"
	userHelper "github.com/slaveofcode/hansip/utils/user"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type FileOpenParam struct {
	Password     string `json:"downloadPassword" binding:"omitempty"`
	UserPassword string `json:"accountPassword" binding:"omitempty"`
}

func Download(pgRepo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		code := c.Param("code")

		var bodyParams FileOpenParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		db := pgRepo.GetDB()

		var shortLink models.ShortLink
		res := db.Preload("FileGroup").Where(`"shortCode" = ?`, code).First(&shortLink)
		if res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Unknown file",
			})
			return
		}

		isProtected := shortLink.PIN != ""
		if isProtected {
			err := bcrypt.CompareHashAndPassword([]byte(shortLink.PIN), []byte(bodyParams.Password))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "Invalid Password" + err.Error(),
				})
				return
			}
		}

		// checking file exist on FS
		if _, err := os.Stat(shortLink.FileGroup.FileKey); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "File not found",
			})
			return
		}

		userId := ""
		isAllowedToDownload := true
		isPrivateMembersOnly := len(shortLink.FileGroup.SharedToUserIds) > 0
		if isPrivateMembersOnly {
			isAllowedToDownload = false

			tokenHeader := c.Request.Header["Authorization"]
			if len(tokenHeader) > 0 {
				user, err := userHelper.GetUserFromHeaderAuth(pgRepo, tokenHeader[0])
				if err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"message": "Please login to continue download" + err.Error(),
					})
					return
				}

				for _, uid := range shortLink.FileGroup.SharedToUserIds {
					if uid == user.ID.String() {
						isAllowedToDownload = true
						userId = user.ID.String()
						break
					}
				}
			}
		}

		if !isAllowedToDownload {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "You're not allowed to download this file",
			})
			return
		}

		if userId != "" {
			var userCred models.UserCredential
			res = db.Where(`"userId" = ?`, userId).First(&userCred)
			err := bcrypt.CompareHashAndPassword([]byte(userCred.CredentialValue), []byte(bodyParams.UserPassword))
			if res.RowsAffected == 0 || err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "Invalid Account Password",
				})
				return
			}
		}

		fileName := fmt.Sprintf("hansip-file-%s.zip", time.Now().Format("20060102150405"))
		filePath := shortLink.FileGroup.FileKey

		if isPrivateMembersOnly {
			var userKey models.UserKey
			resKey := db.Where(`"userId" = ?`, userId).First(&userKey)

			// key not found
			if resKey.Error != nil || resKey.RowsAffected <= 0 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Unable to decrypt file content",
				})
				return
			}

			bundledPath := filepath.FromSlash(viper.GetString("dirpaths.bundle"))
			secretKey := aes.Decrypt(bodyParams.UserPassword, userKey.Private)
			filePathDec, err := age_encryption.DecryptFile(shortLink.FileGroup.FileKey, bundledPath, secretKey)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Unable to decrypt file content" + err.Error(),
				})
				return
			}

			filePath = filePathDec
		}

		c.FileAttachment(filePath, fileName)
	}
}
