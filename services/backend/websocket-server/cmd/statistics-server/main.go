package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/common/logging"
)

func generalStatistics(c *gin.Context) {
	stats, err := business.DbGetStatistics()
	if err != nil {
		c.Error(err)
		c.Status(500)
		return
	}
	c.IndentedJSON(http.StatusOK, stats)
}

func main() {
	logger, err := zap.NewProduction(zap.String("app", "ms-stat"))
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logging.SetLogger(logger)
	// Connect to redis
	business.LoadRedisClient()

	// Serve web server
	logger.Info("init")
	router := gin.Default()

	stats := router.Group("/stats")
	stats.GET("/", generalStatistics)

	router.Run()
}
