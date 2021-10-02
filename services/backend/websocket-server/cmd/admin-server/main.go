package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"mealswipe.app/mealswipe/internal/business"
)

func clearCache(c *gin.Context) {
	params := c.Request.URL.Query()
	if val, ok := params["token"]; ok {
		if len(val) == 0 || val[0] != "skkNW4rdeqYcdsfjDSJFidsjfSDfqwoeokE" {
			c.Status(404)
			log.Println("User gave wrong token, sending them away")
			return
		}
	} else {
		c.Status(404)
		log.Println("User did not provide a token, sending them away")
		return
	}

	cnt, err := business.DbClearCache()
	if err != nil {
		fmt.Println(err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Removed %d from cache but encounted error %s", cnt, err.Error()))
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("Removed %d from cache", cnt))
}

func main() {
	// Connect to redis
	business.LoadRedisClient()

	// Serve web server
	log.Println("Starting...")
	router := gin.Default()

	stats := router.Group("/admin")
	stats.GET("/", clearCache)

	router.Run()
}
