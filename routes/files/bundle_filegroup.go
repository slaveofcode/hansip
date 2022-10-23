package files

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/age_encryption"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
	"github.com/slaveofcode/hansip/routes/middleware"
	appConfig "github.com/slaveofcode/hansip/utils/config"
	"github.com/slaveofcode/hansip/utils/shortlink"
	"github.com/spf13/viper"
	"github.com/yeka/zip"
	"golang.org/x/crypto/bcrypt"
)

type BundleFileGroupParam struct {
	FileGroupId      uint64   `json:"fileGroupId" binding:"required"`
	ExpiredAt        string   `json:"expiredAt" binding:"required,datetime=2006-01-02T15:04:05Z07:00"` // format UTC: 2021-07-18T10:00:00.000Z
	Passcode         string   `json:"passcode" binding:"required,gte=6,lte=100"`
	DownloadPassword string   `json:"downloadPassword" binding:"omitempty,gte=6,lte=100"`
	UserIds          []uint64 `json:"userIds" binding:"omitempty"`
}

func BundleFileGroup(repo repository.Repository, s3Client *s3.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		userId, err := middleware.GetUserId(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Unauthorized request",
			})
			return
		}

		var bodyParams BundleFileGroupParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		passcode, err := bcrypt.GenerateFromPassword([]byte(bodyParams.Passcode), bcrypt.DefaultCost)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Unable to process password",
			})
			return
		}

		db := repo.GetDB()

		var fileGroup models.FileGroup
		res := db.Where(
			`id = ? AND "userId" = ? AND "bundledAt" IS NULL`,
			bodyParams.FileGroupId,
			userId,
		).First(&fileGroup)

		if res.Error != nil || res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Invalid file group",
			})
			return
		}

		var fileItems []models.FileItem
		res = db.Where(`"fileGroupId" = ?`, fileGroup.ID).Find(&fileItems)
		if res.Error != nil || res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Empty file group",
			})
			return
		}

		bundledPath := filepath.FromSlash(viper.GetString("dirpaths.bundle"))
		zipFileName := fmt.Sprintf("%d.zip", fileGroup.ID)
		bundledFullPath := filepath.Join(bundledPath, zipFileName)
		zipFile, err := os.Create(bundledFullPath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to bundle files:" + err.Error(),
			})
			return
		}

		userPubKeys := []string{}
		fileGroupUsers := []models.FileGroupUser{}
		if len(bodyParams.UserIds) > 0 {
			// add self user, for owner access
			userShares := bodyParams.UserIds
			fileGroupUsers = append(fileGroupUsers, models.FileGroupUser{
				FileGroupId: fileGroup.ID,
				UserId:      userId,
			})
			userShares = append(userShares, userId)

			var userKeys []models.UserKey
			res := db.Where(`"userId" IN ?`, userShares).Find(&userKeys)
			if res.RowsAffected > 0 {

				for _, key := range userKeys {
					userPubKeys = append(userPubKeys, key.Public)
					fileGroupUsers = append(fileGroupUsers, models.FileGroupUser{
						FileGroupId: fileGroup.ID,
						UserId:      key.UserId,
					})
				}
			}
		}

		if len(fileGroupUsers) > 0 {
			res := db.Create(&fileGroupUsers)
			if res.Error != nil || res.RowsAffected <= 0 {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Unable to save user sharing info:" + err.Error(),
				})
				return
			}
		}

		uploadPath := filepath.FromSlash(viper.GetString("dirpaths.upload"))
		zipCompressor := zip.NewWriter(zipFile)
		for _, item := range fileItems {
			filePath := filepath.Join(uploadPath, item.Filename)

			f, err := os.Open(filePath)
			if err != nil {
				// skip
				log.Println("Error opening file at:", filePath)
				continue
			}

			// add to compression
			w, err := zipCompressor.Encrypt(item.Realname, bodyParams.Passcode, zip.AES256Encryption)
			if err != nil {
				log.Println("Error prepare zip file at:", filePath, err.Error())
				f.Close()
				continue
			}

			_, err = io.Copy(w, f)
			if err != nil {
				log.Println("Error zipping file at:", filePath, err.Error())
				f.Close()
				continue
			}

			f.Close()
			os.Remove(filePath)
		}

		zipCompressor.Flush()
		zipCompressor.Close()

		fileGroup.FileKey = bundledFullPath

		// set age encryption first if user exist
		if len(userPubKeys) > 0 {
			filePathEnc, err := age_encryption.EncryptFile(bundledFullPath, bundledPath, userPubKeys)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Unable to encrypt the file:" + err.Error(),
				})
				return
			}

			fileGroup.FileKey = filePathEnc
			os.Remove(bundledFullPath)
		}

		expDate, err := time.Parse(time.RFC3339, bodyParams.ExpiredAt)
		if err != nil {
			expDate = time.Now().Add(time.Hour * 24 * 30) // 30 days default
		}

		now := time.Now()
		fileGroup.ArchivePasscode = string(passcode)
		fileGroup.BundledAt = &now
		fileGroup.ExpiredAt = &expDate

		res = db.Save(&fileGroup)
		if res.Error != nil || res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to bundle files:" + err.Error(),
			})
			return
		}

		if appConfig.IsUsingS3Storage() {
			go func(filePath string) {
				keyName := filepath.Base(filePath)
				bundledFile, err := os.Open(filePath)
				if err != nil {
					log.Printf("Error reading bundled file at %s, is the file removed? %s", filePath, err.Error())
					return
				}
				defer bundledFile.Close()

				_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
					Bucket:  aws.String(appConfig.GetS3Bucket()),
					Key:     &keyName,
					Body:    bundledFile,
					Expires: fileGroup.ExpiredAt, // cache expiration
				})

				if err == nil {
					fileGroup.FileKey = keyName
					db.Save(&fileGroup)

					// remove local file because already uploaded to S3
					os.Remove(filePath)
					return
				}

				log.Println("error S3 upload", err)
			}(fileGroup.FileKey)
		}

		pinCode := ""

		if len(bodyParams.DownloadPassword) > 0 {
			pinEnc, err := bcrypt.GenerateFromPassword([]byte(bodyParams.DownloadPassword), bcrypt.DefaultCost)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Unable to bundle files:" + err.Error(),
				})
				return
			}

			pinCode = string(pinEnc)
		}

		shortLink, err := shortlink.MakeNewCode(fileGroup.ID, pinCode, db)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to create download link:" + err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"expiredAt":   fileGroup.ExpiredAt,
				"downloadUrl": shortlink.MakeURL(shortLink),
			},
		})
	}
}
