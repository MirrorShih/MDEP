package controller

import (
	"MDEP/services"
	"github.com/gin-gonic/gin"
)

func DetectorList(c *gin.Context) {
	services.GetDetectorList(c)
}
