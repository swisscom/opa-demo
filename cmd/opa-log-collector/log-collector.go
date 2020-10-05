package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main(){
	r := gin.Default()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	r.POST("/logs", func(c *gin.Context) {
		logger.Debugf("received log: %v", c.Request.Body)
	})

	_ = r.Run("0.0.0.0:8182")
}
