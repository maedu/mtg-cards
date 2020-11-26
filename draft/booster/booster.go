package booster

import "github.com/gin-gonic/gin"

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/draft/booster2", handleGetSealed) // TODO rename

}
