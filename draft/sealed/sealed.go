package sealed

import (
	"net/http"

	"bitbucket.org/spinnerweb/cards/booster"
	"github.com/gin-gonic/gin"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/draft/sealed", handleGetSealed)

}

func handleGetSealed(c *gin.Context) {
	setName := c.Query("set")

	boosters := []booster.Booster{}

	for i := 0; i < 6; i++ {
		booster, err := booster.GenerateBooster(setName)
		if err != nil {
			c.Error(err)
			return
		}
		boosters = append(boosters, booster)
	}
	c.JSON(http.StatusOK, boosters)
}
