package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/common/logging"
)

func clearCache(c *gin.Context) {
	logger := logging.Get()
	params := c.Request.URL.Query()
	if val, ok := params["token"]; ok {
		if len(val) == 0 || val[0] != "skkNW4rdeqYcdsfjDSJFidsjfSDfqwoeokE" {
			c.Status(404)
			logger.Info("User gave wrong token, sending them away", zap.Int("return", 404))
			return
		}
	} else {
		c.Status(404)
		logger.Info("User did not provide a token, sending them away", zap.Int("return", 404))
		return
	}

	cnt, err := business.DbClearCache()
	if err != nil {
		logger.Error("encountered an error clearing the cache", zap.Int("return", 500), zap.Error(err))
		c.String(http.StatusInternalServerError, fmt.Sprintf("Removed %d from cache but encounted error %s", cnt, err.Error()))
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("Removed %d from cache", cnt))
}

func main() {
	logger, err := zap.NewProduction(zap.Fields(zap.String("app", "ms-admin")))
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

	stats := router.Group("/admin")
	stats.GET("/", clearCache)

	router.Run()
}
