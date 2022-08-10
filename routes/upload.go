package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slaveofcode/securi/age_encryption"
)

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Unable to process the file:" + err.Error(),
		})
		return
	}
	fileExt := filepath.Ext(file.Filename)

	// Generate random file name for the new uploaded file so it doesn't override the old file with same name
	newFileName := uuid.New().String() + fileExt

	dirPath := filepath.FromSlash(os.Getenv("UPLOAD_DIR_PATH"))
	destPath := filepath.Join(dirPath, newFileName)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Unable to save the file:" + err.Error(),
		})
		return
	}

	// res, err := age_encryption.EncryptFile(destPath, dirPath)
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
	// 		"status":  false,
	// 		"message": "Unable to encrypt the file:" + err.Error(),
	// 	})
	// 	return
	// }

	res, err := age_encryption.DecryptFile(destPath, dirPath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Unable to encrypt the file:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "File uploaded:" + res,
	})
}
