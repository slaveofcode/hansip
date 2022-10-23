package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
)

const (
	LimitResultCount = 25
)

type UserQuery struct {
	Keyword string `form:"q" binding:"omitempty"`
}

type userResults struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

func UserQueries(repo repository.Repository) func(c *gin.Context) {
	return func(c *gin.Context) {
		var query UserQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid query param",
			})
			return
		}

		db := repo.GetDB()

		var users []models.User
		res := db.Where(`"name" LIKE ? OR "alias" LIKE ?`, "%"+query.Keyword+"%", "%"+query.Keyword+"%").
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
				ID:    user.ID,
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
