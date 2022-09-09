package visit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/repository/pg/models"
	userHelper "github.com/slaveofcode/hansip/utils/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type FileResponse struct {
	FileId   string               `json:"fileId"`
	FileName string               `json:"fileName"`
	FileType models.PreviewAsType `json:"fileType"`
	FileSize int64                `json:"fileSize"`
}

type FileOpenParam struct {
	Password     string `json:"downloadPassword" binding:"omitempty"`
	UserPassword string `json:"accountPassword" binding:"omitempty"`
}

func getFiles(db *gorm.DB, fileGroupId *uuid.UUID) ([]FileResponse, error) {
	files := []FileResponse{}

	var fileItems []models.FileItem
	res := db.Where(`"fileGroupId" = ?`, fileGroupId.String()).Find(&fileItems)
	if res.RowsAffected <= 0 {
		return files, nil
	}

	for _, fileItem := range fileItems {
		files = append(files, FileResponse{
			FileId:   fileItem.ID.String(),
			FileName: fileItem.Realname,
			FileType: fileItem.PreviewAs,
			FileSize: fileItem.SizeInBytes,
		})
	}

	return files, nil
}

func View(pgRepo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		code := c.Param("code")

		// get the detail shortlink
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
		files := []FileResponse{}
		if !isProtected {
			files, _ = getFiles(db, shortLink.FileGroupId)
		}

		isNeedLogin := len(shortLink.FileGroup.SharedToUserIds) > 0

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"isProtected": isProtected,
			"isNeedLogin": isNeedLogin,
			"data": gin.H{
				"files": files,
			},
		})
	}
}

func ViewProtected(pgRepo *pg.RepositoryPostgres) func(c *gin.Context) {
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

		err := bcrypt.CompareHashAndPassword([]byte(shortLink.PIN), []byte(bodyParams.Password))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Invalid Page Password" + err.Error(),
			})
			return
		}

		isAllowedToOpen := false
		isNeedLogin := len(shortLink.FileGroup.SharedToUserIds) > 0
		var user *models.User = nil
		if isNeedLogin {
			tokenHeader := c.Request.Header["Authorization"]
			if len(tokenHeader) > 0 {
				user, err = userHelper.GetUserFromHeaderAuth(pgRepo, tokenHeader[0])
				if err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"message": "Please login to continue download" + err.Error(),
					})
					return
				}

				for _, uid := range shortLink.FileGroup.SharedToUserIds {
					if uid == user.ID.String() {
						isAllowedToOpen = true
						break
					}
				}
			}
		}

		if !isAllowedToOpen {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "You're not allowed to download",
			})
			return
		}

		var userCred models.UserCredential
		res = db.Where(`"userId" = ?`, user.ID.String()).First(&userCred)
		err = bcrypt.CompareHashAndPassword([]byte(userCred.CredentialValue), []byte(bodyParams.UserPassword))
		if res.RowsAffected == 0 || err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Invalid Account Password",
			})
			return
		}

		files, _ := getFiles(db, shortLink.FileGroupId)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"files": files,
			},
		})
	}
}
