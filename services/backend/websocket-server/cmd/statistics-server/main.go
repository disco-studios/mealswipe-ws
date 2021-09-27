package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"mealswipe.app/mealswipe/internal/business"
)

func generalStatistics(c *gin.Context) {
	log.Println("> Hit!")

	stats, err := business.DbGetStatistics()
	if err != nil {
		c.Error(err)
		c.Status(500)
		return
	}
	c.IndentedJSON(http.StatusOK, stats)
}

func main() {
	// Connect to redis
	business.LoadRedisClient()

	// Serve web server
	log.Println("Starting...")
	router := gin.Default()

	stats := router.Group("/stats")
	stats.GET("/", generalStatistics)

	router.Run()
}
