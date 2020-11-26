package server

import (

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Configure() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	/*r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true //return strings.Contains(origin, "localhost")
		},
		MaxAge: 12 * time.Hour,
	}))*/

	return r
}
