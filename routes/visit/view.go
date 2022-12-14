package visit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
	userHelper "github.com/slaveofcode/hansip/utils/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type FileResponse struct {
	FileId   uint64               `json:"fileId"`
	FileName string               `json:"fileName"`
	FileType models.PreviewAsType `json:"fileType"`
	FileSize int64                `json:"fileSize"`
}

type FileOpenParam struct {
	Password     string `json:"downloadPassword" binding:"omitempty"`
	UserPassword string `json:"accountPassword" binding:"omitempty"`
}

func getFiles(db *gorm.DB, fileGroupId uint64) ([]FileResponse, error) {
	files := []FileResponse{}

	var fileItems []models.FileItem
	res := db.Where(`"fileGroupId" = ?`, fileGroupId).Find(&fileItems)
	if res.RowsAffected <= 0 {
		return files, nil
	}

	for _, fileItem := range fileItems {
		files = append(files, FileResponse{
			FileId:   fileItem.ID,
			FileName: fileItem.Realname,
			FileType: fileItem.PreviewAs,
			FileSize: fileItem.SizeInBytes,
		})
	}

	return files, nil
}

func View(pgRepo repository.Repository) func(c *gin.Context) {
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

		var fileGroupUsers []models.FileGroupUser
		res = db.Where(`"fileGroupId" = ?`, shortLink.FileGroupId).Find(&fileGroupUsers)
		if res.Error != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Failed getting user sharing info:" + res.Error.Error(),
			})
			return
		}

		isAllowedToOpen := true
		isNeedLogin := len(fileGroupUsers) > 0
		if isNeedLogin {
			isAllowedToOpen = false
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

				for _, fgu := range fileGroupUsers {
					if fgu.UserId == user.ID {
						isAllowedToOpen = true
						break
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"isProtected": isProtected,
			"isNeedLogin": isNeedLogin,
			"isAllowed":   isAllowedToOpen,
			"data": gin.H{
				"files": files,
			},
		})
	}
}

func ViewProtected(repo repository.Repository) func(c *gin.Context) {
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

		if shortLink.PIN != "" {
			err := bcrypt.CompareHashAndPassword([]byte(shortLink.PIN), []byte(bodyParams.Password))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "Invalid Page Password" + err.Error(),
				})
				return
			}
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

		isAllowedToOpen := true
		isNeedLogin := len(fileGroupUsers) > 0
		var user *models.User = nil
		if isNeedLogin {
			isAllowedToOpen = false
			tokenHeader := c.Request.Header["Authorization"]
			if len(tokenHeader) > 0 {
				userAuth, err := userHelper.GetUserFromHeaderAuth(repo, tokenHeader[0])
				if err != nil {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"message": "Please login to continue download" + err.Error(),
					})
					return
				}

				user = userAuth

				for _, fgu := range fileGroupUsers {
					if fgu.UserId == user.ID {
						isAllowedToOpen = true
						break
					}
				}
			}
		}

		if !isAllowedToOpen {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "You're not allowed to view this page",
			})
			return
		}

		var userCred models.UserCredential
		res = db.Where(`"userId" = ?`, user.ID).First(&userCred)
		err := bcrypt.CompareHashAndPassword([]byte(userCred.CredentialValue), []byte(bodyParams.UserPassword))
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
