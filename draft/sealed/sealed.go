package sealed

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/booster"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/draft/sealed", handleGetSealed)

}

func handleGetSealed(c *gin.Context) {
	setNames := c.QueryArray("set")
	boosterType := c.Query("type")
	if boosterType == "" {
		boosterType = booster.Commander
	}

	boosters, err := booster.GenerateBoosters(boosterType, setNames)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, boosters)
}
