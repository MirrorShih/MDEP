package controller

import (
	"MDEP/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetDetectorList(c *gin.Context) {
	projection := bson.D{{"detector_id", 1}, {"detector_name", 1}}
	results := services.MongoClient.ListDetector("MDEP", "detector", projection)
	var response []bson.M
	for _, res := range results {
		response = append(response, bson.M{"detector_id": res.Id.Hex(), "detector_name": res.Name})
	}
	c.JSON(http.StatusOK, response)
}

func CreateDetector(c *gin.Context) {
	c.JSON(http.StatusNoContent, gin.H{})
}

func CreateTask(c *gin.Context) {

}

func GetReportList(c *gin.Context) {

}

func GetReport(c *gin.Context) {

}
func UpdateDetector(c *gin.Context) {

}

func DeleteDetector(c *gin.Context) {

}
