package files

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/repository/pg/models"
	"github.com/slaveofcode/securi/routes/middleware"
	"golang.org/x/crypto/bcrypt"
)

type FileGroupParam struct {
	ArchiveType      models.ArchiveType `json:"archiveType" binding:"required"`
	Passcode         string             `json:"passcode" binding:"required,gte=6,lte=100"`
	DownloadPassword string             `json:"downloadPassword" binding:"omitempty,gte=6,lte=100"`
}

func CreateFileGroup(repo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		var bodyParams FileGroupParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		userId, err := uuid.Parse(c.GetString(middleware.CTX_USER_ID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		db := repo.GetDB()
		passcode, err := bcrypt.GenerateFromPassword([]byte(bodyParams.Passcode), bcrypt.DefaultCost)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		fg := models.FileGroup{
			UserId:                &userId,
			ArchiveType:           bodyParams.ArchiveType,
			ArchivePasscode:       string(passcode),
			MaxDownload:           0,
			DeleteAtDownloadTimes: 0,
			TotalFiles:            0,
		}

		res := db.Create(&fg)
		if res.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Unable prepare filegroup upload",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": true,
			"data": gin.H{
				"fgId": fg.ID,
			},
		})
	}
}
