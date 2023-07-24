package controller

import (
	"MDEP/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	c.JSON(http.StatusNoContent, gin.H{})
}

func GetReportList(c *gin.Context) {

}

func GetReport(c *gin.Context) {
	target := c.Param("report_id")
	id, _ := primitive.ObjectIDFromHex(target)
	filter := bson.D{bson.E{Key: "_id", Value: id}}
	result := services.MongoClient.GetCertainReport("MDEP", "report", filter)
	c.JSON(http.StatusOK, gin.H{"report_id": result.Id.Hex(), "function_type": result.FuncType,
		"accuracy": result.Accuracy, "false_positive": result.FP,
		"false_negative": result.FN, "precision": result.Precision,
		"recall": result.Recall, "f1_score": result.F1,
		"testing_time": result.TestTime, "testing_sample_num": result.TestSampleNum,
		"total_sample_num": result.TotalSampleNum})
}

func UpdateDetector(c *gin.Context) {

}

func DeleteDetector(c *gin.Context) {
	c.JSON(http.StatusNoContent, gin.H{})
}
