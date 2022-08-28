package files

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/repository/pg/models"
	"github.com/slaveofcode/securi/routes/middleware"
)

type FileGroupParam struct {
	ArchiveType models.ArchiveType `json:"archiveType" binding:"required"`
}

func CreateFileGroup(repo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		userId, err := middleware.GetUserId(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Unauthorized request",
			})
			return
		}

		var bodyParams FileGroupParam
		if err := c.ShouldBindJSON(&bodyParams); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid body request",
			})
			return
		}

		db := repo.GetDB()

		fg := models.FileGroup{
			UserId:                &userId,
			ArchiveType:           bodyParams.ArchiveType,
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
			"success": true,
			"data": gin.H{
				"fgId": fg.ID,
			},
		})
	}
}
