package files

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/repository/pg/models"
	"github.com/slaveofcode/securi/routes/middleware"
	"github.com/slaveofcode/securi/utils/shortlink"
	"github.com/yeka/zip"
	"golang.org/x/crypto/bcrypt"
)

type BundleFileGroupParam struct {
	FileGroupId      uuid.UUID `json:"fileGroupId" binding:"required"`
	ExpiredAt        string    `json:"expiredAt" binding:"required,datetime=2006-01-02T15:04:05Z07:00"` // format UTC: 2021-07-18T10:00:00.000Z
	Passcode         string    `json:"passcode" binding:"required,gte=6,lte=100"`
	DownloadPassword string    `json:"downloadPassword" binding:"omitempty,gte=6,lte=100"`
}

func BundleFileGroup(repo *pg.RepositoryPostgres) func(c *gin.Context) {
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

		var fg models.FileGroup
		res := db.Where(
			`id = ? AND "userId" = ? AND "bundledAt" IS NULL`,
			bodyParams.FileGroupId.String(),
			userId.String()).First(&fg)

		if res.Error != nil || res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Invalid file group",
			})
			return
		}

		var fileItems []models.FileItem
		res = db.Where(`"fileGroupId" = ?`, fg.ID.String()).Find(&fileItems)
		if res.Error != nil || res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Empty file group",
			})
			return
		}

		bundledPath := filepath.FromSlash(os.Getenv("BUNDLED_DIR_PATH"))
		bundledFullPath := filepath.Join(bundledPath, fg.ID.String()+".zip")
		zipFile, err := os.Create(bundledFullPath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to bundle files:" + err.Error(),
			})
			return
		}

		uploadPath := filepath.FromSlash(os.Getenv("UPLOAD_DIR_PATH"))
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

		expDate, err := time.Parse(time.RFC3339, bodyParams.ExpiredAt)
		if err != nil {
			expDate = time.Now().Add(time.Hour * 24 * 30) // 30 days default
		}

		now := time.Now()
		fg.ArchivePasscode = string(passcode)
		fg.BundledAt = &now
		fg.ExpiredAt = &expDate

		res = db.Save(&fg)
		if res.Error != nil || res.RowsAffected <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to bundle files:" + err.Error(),
			})
			return
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

		shortLink, err := shortlink.MakeNewCode(&fg.ID, pinCode, db)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to create download link:" + err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": true,
			"data": gin.H{
				"expiredAt":   fg.ExpiredAt,
				"downloadUrl": shortlink.MakeURL(shortLink),
			},
		})
	}
}
