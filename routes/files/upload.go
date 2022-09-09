package files

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/repository/pg/models"
	fileHelper "github.com/slaveofcode/hansip/utils/file"
	"gorm.io/gorm"
)

func Upload(repo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		fileGroupId, err := uuid.Parse(c.PostForm("fileGroupId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid group file:" + err.Error(),
			})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Unable to process the file:" + err.Error(),
			})
			return
		}

		fileExt := filepath.Ext(file.Filename)

		// Generate random file name for the new uploaded file so it doesn't override the old file with same name
		newFileName := uuid.New().String() + fileExt

		uploadPath := filepath.FromSlash(os.Getenv("UPLOAD_DIR_PATH"))
		uploadFullPath := filepath.Join(uploadPath, newFileName)
		if err := c.SaveUploadedFile(file, uploadFullPath); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable to save the file:" + err.Error(),
			})
			return
		}

		db := repo.GetDB()

		savedFile, _ := os.Open(uploadFullPath)
		filePreview := fileHelper.GetHeadFilePreviewValue(savedFile)
		fileSize, _ := os.Stat(uploadFullPath)

		saveFileMeta := func() error {
			return db.Transaction(func(tx *gorm.DB) error {
				var fileGroup models.FileGroup
				res := tx.Where("id = ?", fileGroupId.String()).First(&fileGroup)

				if res.Error != nil {
					return res.Error
				}

				fileItem := models.FileItem{
					FileGroupId: &fileGroupId,
					Filename:    newFileName,
					Realname:    file.Filename,
					PreviewAs:   filePreview,
					SizeInBytes: fileSize.Size(),
				}
				res = tx.Create(&fileItem)
				if res.Error != nil {
					return res.Error
				}

				fileGroup.TotalFiles += 1
				res = tx.Save(fileGroup)
				return res.Error
			}, &sql.TxOptions{Isolation: sql.LevelSerializable})
		}

		for {
			err = saveFileMeta()
			// TODO: Need better recognition for handling serialize transaction error
			serializeErrStr := "ERROR: could not serialize access due to concurrent update (SQLSTATE 40001)"
			if err != nil && err.Error() != serializeErrStr {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Unable to process the file:" + err.Error(),
				})
				return
			}

			if err != nil {
				time.Sleep(time.Millisecond * 10)
			} else {
				break
			}
		}

		// res, err := age_encryption.EncryptFile(destPath, dirPath)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		// 		"success":  false,
		// 		"message": "Unable to encrypt the file:" + err.Error(),
		// 	})
		// 	return
		// }

		// res, err := age_encryption.DecryptFile(destPath, dirPath)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		// 		"success":  false,
		// 		"message": "Unable to encrypt the file:" + err.Error(),
		// 	})
		// 	return
		// }

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "File uploaded",
		})
	}
}
