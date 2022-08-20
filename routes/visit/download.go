package visit

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/securi/repository/pg"
	"github.com/slaveofcode/securi/repository/pg/models"
	"golang.org/x/crypto/bcrypt"
)

type DownloadQuery struct {
	CheckOnly string `form:"check" binding:"omitempty"`
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
				"status":  false,
				"message": "Unknown file",
			})
			return
		}

		isProtected := shortLink.PIN != ""
		if isProtected {
			err := bcrypt.CompareHashAndPassword([]byte(shortLink.PIN), []byte(bodyParams.Password))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status":  false,
					"message": "Invalid Password" + err.Error(),
				})
				return
			}
		}

		// checking file exist on FS
		bundledPath := filepath.FromSlash(os.Getenv("BUNDLED_DIR_PATH"))
		bundledFullPath := filepath.Join(bundledPath, shortLink.FileGroupId.String()+".zip")
		if _, err := os.Stat(bundledFullPath); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "File not found",
			})
			return
		}

		var query DownloadQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid query param",
			})
			return
		}

		if query.CheckOnly != "" {
			check, _ := strconv.ParseBool(query.CheckOnly)
			if check {
				c.JSON(http.StatusOK, gin.H{
					"status":  true,
					"message": "File Ready",
				})
				return
			}
		}

		// flush content-disposition attachment
		downloadFileName := fmt.Sprintf("securi-file-%s.zip", time.Now().Format("20060102150405"))
		c.FileAttachment(bundledFullPath, downloadFileName)
	}
}
