package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository/pg"
	"github.com/slaveofcode/hansip/repository/pg/models"
)

const (
	LimitResultCount = 25
)

type UserQuery struct {
	Keyword string `form:"q" binding:"omitempty"`
}

type userResults struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

func UserQueries(pgRepo *pg.RepositoryPostgres) func(c *gin.Context) {
	return func(c *gin.Context) {
		var query UserQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid query param",
			})
			return
		}

		db := pgRepo.GetDB()

		var users []models.User
		res := db.Where(`"name" ILIKE ? OR "alias" ILIKE ?`, "%"+query.Keyword+"%", "%"+query.Keyword+"%").
			Order(`"name" ASC`).
			Limit(LimitResultCount).
			Offset(0).
			Find(&users)

		if res.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"success": false,
				"message": "Unable fetch users",
			})
			return
		}

		results := []userResults{}

		for _, user := range users {
			results = append(results, userResults{
				ID:    user.ID.String(),
				Name:  user.Name,
				Alias: user.Alias,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"users": results,
			},
		})
	}
}
