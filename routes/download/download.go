package download

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/age_encryption"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
	"github.com/slaveofcode/hansip/utils/aes"
	"github.com/slaveofcode/hansip/utils/config"
	userHelper "github.com/slaveofcode/hansip/utils/user"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

var downloadManager *manager.Downloader

type FileOpenParam struct {
	Password     string `json:"downloadPassword" binding:"omitempty"`
	UserPassword string `json:"accountPassword" binding:"omitempty"`
}

func Download(repo repository.Repository, s3Client *s3.Client) func(c *gin.Context) {

	if config.IsUsingS3Storage() && downloadManager == nil {
		downloadManager = manager.NewDownloader(s3Client)
	}

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

		db := repo.GetDB()

		var shortLink models.ShortLink
		res := db.Preload("FileGroup").Where(`"shortCode" = ?`, code).First(&shortLink)
		if res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Unknown file",
			})
			return
		}

		var fileGroupUsers []models.FileGroupUser
		res = db.Where(`"fileGroupId" = ?`, shortLink.FileGroupId).Find(&fileGroupUsers)
		if res.Error != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Failed getting user sharing info:" + res.Error.Error(),
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
			// checking file exist on S3
			if !config.IsUsingS3Storage() {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "File not found",
				})
				return
			}

			_, err := s3Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
				Bucket: aws.String(config.GetS3Bucket()),
				Key:    aws.String(shortLink.FileGroup.FileKey),
			})

			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "File not found",
				})
				return
			}
		}

		var userId uint64
		isAllowedToDownload := true
		isPrivateMembersOnly := len(fileGroupUsers) > 0
		if isPrivateMembersOnly {
			isAllowedToDownload = false

			tokenHeader := c.Request.Header["Authorization"]
			if len(tokenHeader) > 0 {
				user, err := userHelper.GetUserFromHeaderAuth(repo, tokenHeader[0])
				if err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"message": "Please login to continue download" + err.Error(),
					})
					return
				}

				for _, fgu := range fileGroupUsers {
					if fgu.UserId == user.ID {
						isAllowedToDownload = true
						userId = user.ID
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

		if userId != 0 {
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
			storedPath := shortLink.FileGroup.FileKey

			if config.IsUsingS3Storage() {
				downloadedPath, err := pullFileFromS3(storedPath, bundledPath)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"success": false,
						"message": "Unable to get the file",
					})
					return
				}

				storedPath = downloadedPath
			}

			secretKey := aes.Decrypt(bodyParams.UserPassword, userKey.Private)
			decryptedFilePath, err := age_encryption.DecryptFile(storedPath, bundledPath, secretKey)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Unable to decrypt file content" + err.Error(),
				})
				return
			}

			filePath = decryptedFilePath
		}

		c.FileAttachment(filePath, fileName)

		// TODO: Set timer remove downloaded files from S3
	}
}

func pullFileFromS3(fileKey, location string) (string, error) {
	fileDest := filepath.Join(location, fileKey)
	if err := os.MkdirAll(filepath.Dir(fileDest), 0775); err != nil {
		return "", err
	}

	fileBuff, err := os.Create(fileDest)
	if err != nil {
		return "", err
	}
	defer fileBuff.Close()

	_, err = downloadManager.Download(
		context.TODO(),
		fileBuff,
		&s3.GetObjectInput{
			Bucket: aws.String(config.GetS3Bucket()),
			Key:    &fileKey,
		},
	)

	if err != nil {
		return "", err
	}

	return fileDest, nil
}
